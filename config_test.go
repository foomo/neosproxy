package neosproxy

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func loadConfig(t *testing.T) *Config {
	ConfigSetDefaults()
	viper.SetConfigName("config-example")
	conf, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	return conf
}

func TestReadConfiguration(t *testing.T) {
	ConfigSetDefaults()
	viper.SetConfigName("config-example")
	err := readConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadNonExistingConfiguration(t *testing.T) {
	ConfigSetDefaults()
	viper.SetConfigName("config-404")
	err := readConfig()
	if err == nil {
		t.Fatal("did not fail on reading a non existing config file")
	}
}

func TestProxyAddress(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, "1.2.3.4:80", c.Proxy.Address)
}

func TestNeosHost(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, "http://cms-example-hostname/", c.Neos.URL.String())
}

func TestCacheAutoUpdate(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, "15m", c.Cache.AutoUpdateDuration)
}

func TestCallbacks(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, 4, len(c.Callbacks.NotifyOnUpdateHooks))

	callback := c.Callbacks.NotifyOnUpdateHooks[0]
	assert.Equal(t, "stage", callback.Workspace)
	assert.Equal(t, "foomo-stage", callback.Channel)
}

func TestChannels(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, 4, len(c.Channels))

	channel, ok := c.Channels["foomo-stage"]
	if !ok {
		t.Fatal("foomo-stage channel not configured")
	}

	assert.Equal(t, "host.example.com", channel.URL.Host)
	assert.Equal(t, "/whatever/to-call", channel.URL.Path)
	assert.Equal(t, "https", channel.URL.Scheme)
	assert.Equal(t, true, channel.VerifyTLS)
	assert.Equal(t, "1234", channel.APIKey)
	assert.Equal(t, "POST", channel.Method)
}

func TestFoomoChannelConfig(t *testing.T) {
	c := loadConfig(t)

	channel, ok := c.Channels["foomo-stage"]
	if !ok {
		t.Fatal("foomo channel not configured")
	}

	assert.Equal(t, ChannelTypeFoomo, channel.Type)
	assert.Equal(t, ChannelTypeFoomo, channel.GetChannelType())
}
func TestSlackChannelConfig(t *testing.T) {
	c := loadConfig(t)

	channel, ok := c.Channels["slack"]
	if !ok {
		t.Fatal("slack channel not configured")
	}

	assert.Equal(t, ChannelTypeSlack, channel.Type)
	assert.Equal(t, ChannelTypeSlack, channel.GetChannelType())
}

func TestGetWorkspaceForChannelIdentifier(t *testing.T) {
	c := loadConfig(t)

	workspace, err := c.GetWorkspaceForChannelIdentifier("foomo-stage")
	assert.NoError(t, err, "no error expected")
	assert.Equal(t, "stage", workspace, "stage workspace expected")
	assert.NotEqual(t, "live", workspace, "live workspace not expected")

	workspace, err = c.GetWorkspaceForChannelIdentifier("foomo-prod")
	assert.NoError(t, err, "no error expected")
	assert.Equal(t, "live", workspace, "live workspace expected")
	assert.NotEqual(t, "stage", workspace, "stage workspace not expected")
}
