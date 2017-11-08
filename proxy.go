package neosproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type Proxy struct {
	Config                    *Config
	APIKey                    string
	CacheInvalidationChannels map[string](chan time.Time)
}

func (p *Proxy) Run() error {
	p.addInvalidationChannel(DefaultWorkspace)
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
func (p *Proxy) addInvalidationChannel(workspace string) chan time.Time {
	if _, ok := p.CacheInvalidationChannels[workspace]; !ok {
		channel := make(chan time.Time, 1)
		p.CacheInvalidationChannels[workspace] = channel
		go func(workspace string, channel chan time.Time) {
			for {
				sleepTime := 5 * time.Second
				time.Sleep(sleepTime)
				requestTime := <-channel
				if err := p.cacheNeosContentServerExport(workspace); err != nil {
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
	return p.CacheInvalidationChannels[workspace]
}
