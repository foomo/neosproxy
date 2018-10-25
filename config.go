package neosproxy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/viper"
)

type Hook struct {
	Workspace string
	Channel   string
}

type ChannelType string

const (
	ChannelTypeFoomo   ChannelType = "foomo"
	ChannelTypeSlack   ChannelType = "slack"
	ChannelTypeDefault ChannelType = "default"
)

type ChannelInterface interface {
	GetChannelType() ChannelType
}

type Channel struct {
	Type ChannelType
	URL  *url.URL

	VerifyTLS    bool
	APIKey       string
	Method       string
	SlackChannel string
}

func (c *Channel) GetChannelType() ChannelType {
	return c.Type
}

type Config struct {
	Proxy struct {
		Address string
	}
	Neos struct {
		URL *url.URL
	}
	Cache struct {
		AutoUpdateDuration string
		Directory          string
	}
	Channels  map[string]*Channel
	Callbacks struct {
		NotifyOnUpdateHooks []*Hook
	}
}

func (c *Config) GetWorkspaceForChannelIdentifier(channelIdentifier string) (workspace string, e error) {
	for _, hook := range c.Callbacks.NotifyOnUpdateHooks {
		if hook.Channel == channelIdentifier {
			// channel
			channel, ok := c.Channels[hook.Channel]
			if !ok {
				e = fmt.Errorf("channel %s not configured for callback %s", hook.Channel, channelIdentifier)
				return
			}

			if channel.GetChannelType() != ChannelTypeFoomo {
				e = fmt.Errorf("unexpected channel type %s for callback %s", channel.GetChannelType(), channelIdentifier)
				return
			}

			workspace = hook.Workspace
			return
		}
	}

	e = fmt.Errorf("callback %s not configured", channelIdentifier)
	return
}

// ConfigSetDefaults sets the defaults for a config
func ConfigSetDefaults() {
	// default flags
	viper.SetDefault("setdefaults", true)
	viper.SetDefault("proxy.address", "0.0.0.0:80")
	viper.SetDefault("neos.host", "http://neos/")

	// update flags
	viper.SetDefault("cache.autoUpdateDuration", "")
	viper.SetDefault("cache.directoy", os.TempDir())

	// config dir setup
	viper.SetConfigName("config")          // name of config file (without extension)
	viper.AddConfigPath("/etc/neosproxy/") // path to look for the config file in
	viper.AddConfigPath(".")               // optionally look for config in the working directory
}

// read configuration file
func readConfig() error {
	if !viper.GetBool("setdefaults") {
		ConfigSetDefaults()
	}
	return viper.ReadInConfig()
}

// GetConfig reads config file and returns a new configuration
func GetConfig() (config *Config, err error) {

	if readConfig() != nil {
		// log.Fatal(fmt.Errorf("fatal error read config file: %s", configReadErr))
		return
	}

	// parse neos host from config
	neosURL, err := url.Parse(viper.GetString("neos.host"))
	if err != nil {
		return
	}

	// prepare config vo
	config = &Config{}
	config.Proxy.Address = viper.GetString("proxy.address")
	config.Neos.URL = neosURL
	config.Cache.Directory = viper.GetString("cache.directory")
	config.Cache.AutoUpdateDuration = viper.GetString("cache.autoUpdateDuration")
	config.Callbacks.NotifyOnUpdateHooks = make([]*Hook, 0)

	// read channel config
	config.Channels = make(map[string]*Channel, 0)
	var channels = viper.Get("channels")
	if channels == nil {
		err = errors.New("no channel config found")
		return
	}
	for channelName, channelObject := range channels.(map[string]interface{}) {
		for _, channelConfig := range channelObject.([]interface{}) {

			// interface map
			channelMap, ok := channelConfig.(map[string]interface{})
			if !ok {
				log.Println("unable to parse channel")
				continue
			}

			// url / endpoint
			host, ok := channelMap["url"]
			if !ok {
				log.Println("unable to read host config for channel", channelName)
				continue
			}
			hostURL, hostURLErr := url.Parse(host.(string))
			if hostURLErr != nil {
				log.Println("unable to parse host config for channel", channelName, hostURLErr)
				continue
			}

			// tls verification
			verifyTLS, ok := channelMap["verify-tls"]
			if !ok {
				verifyTLS = true
			}

			// api key
			key, ok := channelMap["key"]
			if !ok {
				key = ""
			}

			// http method
			method, ok := channelMap["method"]
			if !ok {
				method = http.MethodPost
			}

			// slack channel name
			slackChannelName, ok := channelMap["channel"]
			if !ok {
				slackChannelName = ""
			}

			// typed channel config
			var channel *Channel
			channelTypeToSwitch, _ := channelMap["type"]

			switch channelTypeToSwitch {
			case string(ChannelTypeFoomo):
				channel = &Channel{
					Type: ChannelTypeFoomo,
					URL:  hostURL,

					VerifyTLS: verifyTLS.(bool),
					APIKey:    key.(string),
					Method:    http.MethodPost,
				}
				break
			case string(ChannelTypeSlack):
				channel = &Channel{
					Type: ChannelTypeSlack,
					URL:  hostURL,

					SlackChannel: slackChannelName.(string),
				}
				break
			default:
				channel = &Channel{
					Type: ChannelTypeDefault,
					URL:  hostURL,

					VerifyTLS: verifyTLS.(bool),
					APIKey:    key.(string),
					Method:    method.(string),
				}
			}

			if channel == nil {
				log.Println("channel config skipped")
				continue
			}

			config.Channels[channelName] = channel
		}
	}

	// read callback config
	var callbacks = viper.Get("callbacks")
	for key, callback := range callbacks.(map[string]interface{}) {

		if key == "notifyonupdate" {
			for _, hook := range callback.([]interface{}) {

				hookMap, ok := hook.(map[string]interface{})
				if !ok {
					log.Println("unable to parse notify on update hook")
					continue
				}

				workspace, ok := hookMap["workspace"]
				if !ok {
					workspace = DefaultWorkspace
				}

				channel, ok := hookMap["channel"]
				if !ok {
					log.Println("unable to read channel name from callback config")
					continue
				}

				_, channelOK := config.Channels[channel.(string)]
				if !channelOK {
					log.Println("channel not defined", channel)
					continue
				}

				hook := &Hook{
					Workspace: workspace.(string),
					Channel:   channel.(string),
				}

				config.Callbacks.NotifyOnUpdateHooks = append(config.Callbacks.NotifyOnUpdateHooks, hook)
			}
		}
	}

	return
}
