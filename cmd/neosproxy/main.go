package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/foomo/neosproxy"
)

func main() {

	// load config
	config, configErr := neosproxy.GetConfig()
	if configErr != nil {
		log.Fatalln("failed to read config:", configErr)
		os.Exit(255)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("missing env variable API_KEY")
	}

	flag.Parse()

	// prepare proxy
	p := &neosproxy.Proxy{
		Config: config,
		APIKey: apiKey,
		CacheInvalidationChannels: make(map[string](chan time.Time)),
	}

	// auto update
	if config.Cache.AutoUpdateDuration != "" {
		autoUpdate, err := time.ParseDuration(config.Cache.AutoUpdateDuration)
		if err != nil {
			log.Fatal("invalid auto-update duration value: " + err.Error())
		}
		go func() {
			log.Println(config.Cache.AutoUpdateDuration, "auto update enabled")
			for {
				time.Sleep(autoUpdate)
				log.Println(fmt.Sprintf("auto update: updating %d cache", len(p.CacheInvalidationChannels)))
				for workspace, channel := range p.CacheInvalidationChannels {
					select {
					case channel <- time.Now():
						log.Println(fmt.Sprintf("auto update: added cache invalidation request to queue for '%s' workspace", workspace))
					default:
						log.Println(fmt.Sprintf("auto update: ignored cache invalidation request due to pending invalidation requests for '%s' workspace", workspace))
					}
				}
			}
		}()
	}

	// run proxy
	log.Println("start proxy on", config.Proxy.Address, "for neos", config.Neos.URL.String(), "with cache in directory:", p.Config.Cache.Directory)
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
