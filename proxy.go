package neosproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Proxy struct {
	Config                         *Config
	APIKey                         string
	CacheInvalidationChannels      map[string](chan time.Time)
	cacheInvalidationChannelsMutex *sync.RWMutex
}

func NewProxy(config *Config, apiKey string) *Proxy {
	return &Proxy{
		Config: config,
		APIKey: apiKey,
		CacheInvalidationChannels:      make(map[string](chan time.Time)),
		cacheInvalidationChannelsMutex: &sync.RWMutex{},
	}
}

func (p *Proxy) Run() error {
	p.addInvalidationChannel(DefaultWorkspace, "cron")
	proxyHandler := httputil.NewSingleHostReverseProxy(p.Config.Neos.URL)
	mux := http.NewServeMux()
	mux.Handle("/contentserver/export/", proxyHandler)
	mux.HandleFunc("/contentserverproxy/cache", p.invalidateCache)
	mux.HandleFunc("/contentserver/export", p.serveCachedNeosContentServerExport)
	mux.Handle("/", proxyHandler)

	return http.ListenAndServe(p.Config.Proxy.Address, mux)
}

// error ...
func (p *Proxy) error(w http.ResponseWriter, r *http.Request, code int, msg string) {
	log.Println(fmt.Sprintf("%d\t%s\t%s", code, r.URL, msg))
	w.WriteHeader(code)
}

// addInvalidationChannel adds a new invalidation channel
func (p *Proxy) addInvalidationChannel(workspace string, user string) chan time.Time {

	var channel chan time.Time
	var ok bool

	p.cacheInvalidationChannelsMutex.RLock()
	channel, ok = p.CacheInvalidationChannels[workspace]
	p.cacheInvalidationChannelsMutex.RUnlock()

	if !ok {
		channel = make(chan time.Time, 1)
		p.cacheInvalidationChannelsMutex.Lock()
		p.CacheInvalidationChannels[workspace] = channel
		p.cacheInvalidationChannelsMutex.Unlock()
		go func(workspace string, channel chan time.Time) {
			for {
				sleepTime := 5 * time.Second
				time.Sleep(sleepTime)
				requestTime := <-channel
				if err := p.cacheNeosContentServerExport(workspace, user); err != nil {
					log.Println(err.Error())
				} else {
					log.Println(fmt.Sprintf(
						"processed cache invalidation request, which has been added at %s in %.2fs for workspace %s",
						requestTime.Format(time.RFC3339),
						time.Since(requestTime.Add(sleepTime)).Seconds(),
						workspace,
					))
				}
			}
		}(workspace, channel)
	}

	return channel
}
