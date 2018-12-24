package main

import (
	"flag"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/foomo/neosproxy/cache/content/store/memory"
	"github.com/foomo/neosproxy/client/cms"
	"github.com/foomo/neosproxy/config"
	"github.com/foomo/neosproxy/logging"
	"github.com/foomo/neosproxy/proxy"
)

func main() {

	// parse flags
	flagConfigFile := flag.String("config-file", "/etc/neosproxy/config.yaml", "absolute path to neosproxy config file")
	flag.Parse()

	// logger
	logger := logging.GetDefaultLogEntry()

	// load config
	config, errConfig := config.Load(*flagConfigFile)
	if errConfig != nil {
		logger.WithError(errConfig).Fatalln("failed to Load config")
	}

	if config == nil {
		logger.Fatalln("no config")
	}

	// check token
	if config.Proxy.Token == "" {
		logger.Fatalln("missing config: proxy.token")
	}

	// create cms content load client
	contentLoader, errContentLoader := cms.New(config.Neos.URL.String())
	if errContentLoader != nil {
		logger.WithError(errContentLoader).Fatalln("failed to init cms content loader client")
	}

	// create content cache store
	contentStore := memory.NewCacheStore()
	cacheLifetime := time.Duration(0) // forever // time.Minute * 60

	// create proxy
	p := proxy.New(config, contentLoader.CMS, contentStore, cacheLifetime)

	// // auto update
	// if config.Cache.AutoUpdateDuration != "" {
	// 	autoUpdate, err := time.ParseDuration(config.Cache.AutoUpdateDuration)
	// 	if err != nil {
	// 		log.Fatal("invalid auto-update duration value: " + err.Error())
	// 	}
	// 	go func() {
	// 		log.Println(config.Cache.AutoUpdateDuration, "auto update enabled")
	// 		for {
	// 			time.Sleep(autoUpdate)
	// 			log.Println(fmt.Sprintf("auto update: updating %d cache", len(p.CacheInvalidationChannels)))
	// 			for workspace, channel := range p.CacheInvalidationChannels {
	// 				select {
	// 				case channel <- time.Now():
	// 					log.Println(fmt.Sprintf("auto update: added cache invalidation request to queue for '%s' workspace", workspace))
	// 				default:
	// 					log.Println(fmt.Sprintf("auto update: ignored cache invalidation request due to pending invalidation requests for '%s' workspace", workspace))
	// 				}
	// 			}
	// 		}
	// 	}()
	// }

	// logging
	logger.WithFields(logrus.Fields{
		logging.FieldAddr: config.Proxy.Address,
		"neos":            config.Neos.URL,
		"cache":           config.Cache.Directory,
	}).Info("run proxy server")

	// run proxy
	if err := p.Run(); err != nil {
		logger.WithError(err).Fatalln("failed running proxy server")
	}
}
