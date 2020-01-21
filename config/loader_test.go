package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoader(t *testing.T) {
	cfg, errLoadCfg := Load("../config-example.yaml")
	assert.NoError(t, errLoadCfg)
	assert.NotNil(t, cfg)

	assert.Equal(t, "127.0.0.1:8000", cfg.Proxy.Address)
	assert.Equal(t, "advbfsb-adfgsgsg-4435sgs-afgsgdfg", cfg.Proxy.Token)
	assert.Equal(t, "/neosproxy", cfg.Proxy.BasePath)
	assert.Equal(t, "http", cfg.Neos.URL.Scheme)
	assert.Equal(t, "cms", cfg.Neos.URL.Hostname())

	assert.NotNil(t, cfg.Neos.Dimensions)
	assert.Len(t, cfg.Neos.Dimensions, 2)
	assert.Contains(t, cfg.Neos.Dimensions, "de")
	assert.Contains(t, cfg.Neos.Dimensions, "fr")

	assert.Equal(t, cfg.Neos.Workspace, "live")

	assert.Equal(t, "30m", cfg.Cache.AutoUpdateDuration)
	assert.Equal(t, "/var/data/neosproxy", cfg.Cache.Directory)

	assert.NotNil(t, cfg.Subscriptions)
	assert.Len(t, cfg.Subscriptions, 3)
	assert.Contains(t, cfg.Subscriptions, "foomo-prod")
	assert.Contains(t, cfg.Subscriptions, "foomo-stage")
	assert.Contains(t, cfg.Subscriptions, "slack")

	assert.NotEmpty(t, cfg.Observers)
	assert.Len(t, cfg.Observers, 4)

	assert.NotNil(t, cfg.Observers[0].Foomo)
	assert.Nil(t, cfg.Observers[0].Webhook)
	assert.Nil(t, cfg.Observers[0].Slack)
	assert.Equal(t, ObserverTypeFoomo, cfg.Observers[0].Foomo.Type)
	assert.NotEmpty(t, cfg.Observers[0].Foomo.Name)

	assert.NotNil(t, cfg.Observers[2].Slack)
	assert.Nil(t, cfg.Observers[2].Foomo)
	assert.Nil(t, cfg.Observers[2].Webhook)
	assert.Equal(t, ObserverTypeSlack, cfg.Observers[2].Slack.Type)
	assert.NotEmpty(t, cfg.Observers[2].Slack.Name)

	assert.NotNil(t, cfg.Observers[3].Webhook)
	assert.Nil(t, cfg.Observers[3].Foomo)
	assert.Nil(t, cfg.Observers[3].Slack)
	assert.Equal(t, ObserverTypeWebhook, cfg.Observers[3].Webhook.Type)
	assert.NotEmpty(t, cfg.Observers[3].Webhook.Name)
}
