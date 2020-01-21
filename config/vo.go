package config

import "net/url"

//-----------------------------------------------------------------------------
// ~ Interface
//-----------------------------------------------------------------------------

type ObserverInterface interface {
	GetObserverType() ObserverType
}

//-----------------------------------------------------------------------------
// ~ Public value objects
//-----------------------------------------------------------------------------

// Config struct definition
type Config struct {
	Proxy struct {
		Address  string
		Token    string
		BasePath string
	}
	Neos          Neos
	Cache         Cache
	Observers     []*Observer `json:"-" yaml:"observer"` // *Observer
	Subscriptions []string
}

// Cache config struct
type Cache struct {
	AutoUpdateDuration string `json:"autoUpdateDuration" yaml:"autoUpdateDuration"`
	Directory          string
}

// Neos config struct
type Neos struct {
	URL        *url.URL
	Workspace  string
	Dimensions []string
}

// Observer config struct
type Observer struct {
	Webhook *ObserverWebhook
	Slack   *ObserverSlack
	Foomo   *ObserverFoomo
}

// ObserverSlack struct definition
type ObserverSlack struct {
	Name    string
	Type    ObserverType
	URL     *url.URL
	Channel string
}

// ObserverFoomo struct definition
type ObserverFoomo struct {
	Name      string
	Type      ObserverType
	URL       *url.URL
	VerifyTLS bool
}

// ObserverWebhook struct definition
type ObserverWebhook struct {
	Name      string
	Type      ObserverType
	URL       *url.URL
	VerifyTLS bool
	Token     string
}

//-----------------------------------------------------------------------------
// ~ Private value objects
//-----------------------------------------------------------------------------

type configFile struct {
	Proxy struct {
		Address  string
		Token    string
		BasePath string
	}
	Neos struct {
		URL        string `json:"url" yaml:"url"`
		Workspace  string
		Dimensions []string
	}
	Cache struct {
		AutoUpdateDuration string `json:"autoUpdateDuration" yaml:"autoUpdateDuration"`
		Directory          string
	}
	Observers     []configFileObserver `json:"-" yaml:"observers"`
	Subscriptions []string
}

type configFileObserver struct {
	Name      string
	Type      ObserverType
	URL       string
	VerifyTLS bool   `json:"verify-tls" yaml:"verify-tls"`
	Token     string `json:"token" yaml:"token"`
	Channel   string
}
