package neosproxy

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNotifySlack(t *testing.T) {
	ConfigSetDefaults()
	viper.SetConfigName("config-example")
	config, configErr := GetConfig()
	if configErr != nil {
		t.Fatal(configErr)
	}

	proxy := Proxy{
		APIKey: "1234",
		Config: config,
	}

	channel, ok := config.Channels["slack"]
	if !ok {
		t.Fatal("channel slack not configured")
	}

	success := proxy.notify(channel, "stage", "frederik", NotifierUpdateEvent)
	assert.Equal(t, true, success)
}
