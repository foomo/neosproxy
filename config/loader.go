package config

import (
	"errors"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/foomo/neosproxy/logging"

	yaml "gopkg.in/yaml.v2"
)

//-----------------------------------------------------------------------------
// ~ Public methods
//-----------------------------------------------------------------------------

// Load config from file and return a new configuration
func Load(filename string) (config *Config, err error) {

	// logger
	log := logging.GetDefaultLogEntry().WithField(logging.FieldFunction, "load config")

	// read config file
	data, errReadFile := ioutil.ReadFile(filename)
	if errReadFile != nil {
		err = errReadFile
		return
	}

	// parse config file
	conf := &configFile{}
	errUnmarshal := yaml.Unmarshal(data, &conf)
	if errUnmarshal != nil {
		err = errUnmarshal
		return
	}

	// parse neos host from config
	neosURL, errNeosURLParser := url.Parse(conf.Neos.URL)
	if errNeosURLParser != nil {
		err = errNeosURLParser
		return
	}

	// create config value object
	config = &Config{
		Proxy:         conf.Proxy,
		Cache:         conf.Cache,
		Subscriptions: []string{},
		Observers:     []*Observer{},
	}
	config.Neos.URL = neosURL

	// workspace
	config.Neos.Workspace = conf.Neos.Workspace

	// dimensions
	dimensionNames := map[string]bool{}
	config.Neos.Dimensions = make([]string, len(conf.Neos.Dimensions))
	for index, dimension := range conf.Neos.Dimensions {
		dimension = strings.ToLower(dimension)
		config.Neos.Dimensions[index] = dimension
		dimensionNames[dimension] = true
	}

	// consume and validate observers configuration
	observerNames := map[string]bool{}
	for _, cfgFileObserver := range conf.Observers {

		cfgFileObserver.Name = strings.ToLower(cfgFileObserver.Name)

		var errObserver error
		observer := &Observer{}

		switch cfgFileObserver.Type {
		case ObserverTypeWebhook:
			observer.Webhook, errObserver = newObserverWebhook(cfgFileObserver)
			break
		case ObserverTypeFoomo:
			observer.Foomo, errObserver = newObserverFoomo(cfgFileObserver)
			break
		case ObserverTypeSlack:
			observer.Slack, errObserver = newObserverSlack(cfgFileObserver)
			break
		default:
			errObserver = errors.New("unknown observer type")
		}

		if errObserver != nil {
			log.WithError(errObserver).WithField("observer", cfgFileObserver.Name).Warn("skipped observer: not well configured")
			continue
		}

		observerNames[cfgFileObserver.Name] = true
		config.Observers = append(config.Observers, observer)
	}

	// check if subscriptions match observers
	observers := []string{}
	for _, subscription := range conf.Subscriptions {
		observer := subscription
		if _, ok := observerNames[observer]; !ok {
			log.WithField("observer", observer).Warn("ignore subscription: observer not defined")
			continue
		}
		observers = append(observers, observer)
	}
	config.Subscriptions = observers

	return
}
