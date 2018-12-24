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
		Subscriptions: make(map[string][]string, len(conf.Subscriptions)),
		Observer:      []*Observer{},
	}
	config.Neos.URL = neosURL

	// workspaces
	workspaceNames := map[string]bool{}
	config.Neos.Workspaces = make([]string, len(conf.Neos.Workspaces))
	for index, workspace := range conf.Neos.Workspaces {
		workspace = strings.ToLower(workspace)
		config.Neos.Workspaces[index] = workspace
		workspaceNames[workspace] = true
	}

	// observers
	observerNames := map[string]bool{}
	for _, cfgFileObserver := range conf.Observer {

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
		config.Observer = append(config.Observer, observer)
	}

	// subscriptions
	for workspace, subscriptions := range conf.Subscriptions {
		workspace = strings.ToLower(workspace)
		if _, ok := workspaceNames[workspace]; !ok {
			log.WithField(logging.FieldWorkspace, workspace).Warn("ignore subscriptions: workspace not defined")
			continue
		}

		observers := []string{}
		for _, observer := range subscriptions {
			observer = strings.ToLower(observer)
			if _, ok := observerNames[observer]; !ok {
				log.WithField(logging.FieldWorkspace, workspace).WithField("observer", observer).Warn("ignore subscription: observer not defined")
				continue
			}
			observers = append(observers, observer)
		}
		config.Subscriptions[workspace] = observers
	}

	return
}
