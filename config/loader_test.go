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
	assert.Equal(t, "http", cfg.Neos.URL.Scheme)
	assert.Equal(t, "cms-example-hostname", cfg.Neos.URL.Hostname())

	assert.Len(t, cfg.Neos.Workspaces, 3)
	assert.Equal(t, "live", cfg.Neos.Workspaces[0])
	assert.Contains(t, cfg.Neos.Workspaces, "live")
	assert.Contains(t, cfg.Neos.Workspaces, "stage")
	assert.Contains(t, cfg.Neos.Workspaces, "test")

	assert.Equal(t, "30m", cfg.Cache.AutoUpdateDuration)
	assert.Equal(t, "/tmp/cache", cfg.Cache.Directory)

	assert.Len(t, cfg.Subscriptions, 3)
	assert.Contains(t, cfg.Subscriptions, "live")
	assert.Contains(t, cfg.Subscriptions, "stage")
	assert.Contains(t, cfg.Subscriptions, "test")
	assert.NotNil(t, cfg.Subscriptions["live"])
	assert.NotNil(t, cfg.Subscriptions["stage"])
	assert.NotNil(t, cfg.Subscriptions["test"])

	assert.NotEmpty(t, cfg.Observer)
	assert.Len(t, cfg.Observer, 4)

	assert.NotNil(t, cfg.Observer[0].Foomo)
	assert.Nil(t, cfg.Observer[0].Webhook)
	assert.Nil(t, cfg.Observer[0].Slack)
	assert.Equal(t, ObserverTypeFoomo, cfg.Observer[0].Foomo.Type)
	assert.NotEmpty(t, cfg.Observer[0].Foomo.Name)

	assert.NotNil(t, cfg.Observer[2].Slack)
	assert.Nil(t, cfg.Observer[2].Foomo)
	assert.Nil(t, cfg.Observer[2].Webhook)
	assert.Equal(t, ObserverTypeSlack, cfg.Observer[2].Slack.Type)
	assert.NotEmpty(t, cfg.Observer[2].Slack.Name)

	assert.NotNil(t, cfg.Observer[3].Webhook)
	assert.Nil(t, cfg.Observer[3].Foomo)
	assert.Nil(t, cfg.Observer[3].Slack)
	assert.Equal(t, ObserverTypeWebhook, cfg.Observer[3].Webhook.Type)
	assert.NotEmpty(t, cfg.Observer[3].Webhook.Name)
}
