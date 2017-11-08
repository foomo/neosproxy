package neosproxy

import (
	"log"
	"net/url"
	"os"

	"github.com/spf13/viper"
)

type Hook struct {
	Workspace string
	URL       *url.URL
	VerifyTLS bool
	APIKey    string
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
	Callbacks struct {
		NotifyOnUpdateHooks []*Hook
	}
}

func setDefaultConfig() {
	// default flags
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
	setDefaultConfig()
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

	var callbacks = viper.Get("callbacks")
	for key, callback := range callbacks.(map[string]interface{}) {

		for _, hook := range callback.([]interface{}) {
			if key == "notifyonupdate" {

				hookMap, ok := hook.(map[string]interface{})
				if !ok {
					log.Println("unable to parse notify on update hook")
					continue
				}

				hookWorkspace, ok := hookMap["workspace"]
				if !ok {
					hookWorkspace = DefaultWorkspace
				}
				hookVerifyTls, ok := hookMap["verify-tls"]
				if !ok {
					hookVerifyTls = true
				}
				hookKey, ok := hookMap["key"]
				if !ok {
					hookKey = ""
				}

				hookHost, ok := hookMap["url"]
				hookHostURL, hookHostURLErr := url.Parse(hookHost.(string))
				if hookHostURLErr != nil {
					// log.Fatal("unable to parse hook url", err)
					return nil, hookHostURLErr
				}

				hook := &Hook{
					Workspace: hookWorkspace.(string),
					APIKey:    hookKey.(string),
					VerifyTLS: hookVerifyTls.(bool),
					URL:       hookHostURL,
				}

				config.Callbacks.NotifyOnUpdateHooks = append(config.Callbacks.NotifyOnUpdateHooks, hook)
			}
		}
	}

	return
}
