package neosproxy

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReadConfiguration(t *testing.T) {
	viper.SetConfigName("config-example")
	err := readConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadNonExistingConfiguration(t *testing.T) {
	viper.SetConfigName("config-404")
	err := readConfig()
	if err == nil {
		t.Fatal("did not fail on reading a non existing config file")
	}
}

func loadConfig(t *testing.T) *Config {
	viper.SetConfigName("config-example")
	conf, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	return conf
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

func TestHooks(t *testing.T) {
	c := loadConfig(t)
	assert.Equal(t, 2, len(c.Callbacks.NotifyOnUpdateHooks))

	callback := c.Callbacks.NotifyOnUpdateHooks[0]
	assert.Equal(t, "host.example.com", callback.URL.Host)
	assert.Equal(t, "/whatever/to-call", callback.URL.Path)
	assert.Equal(t, "https", callback.URL.Scheme)

	assert.Equal(t, "1234", callback.APIKey)
	assert.Equal(t, true, callback.VerifyTLS)
}
