package proxy

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/foomo/neosproxy/cache"
	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
	"github.com/foomo/neosproxy/config"
	"github.com/foomo/neosproxy/logging"
	"github.com/foomo/neosproxy/model"
	"github.com/foomo/neosproxy/notifier"
	"github.com/gorilla/mux"

	content_cache "github.com/foomo/neosproxy/cache/content"
)

// New proxy
func New(cfg *config.Config, contentLoader cms.ContentLoader, contentStore store.CacheStore, cacheLifetime time.Duration) *Proxy {

	reverseProxy := httputil.NewSingleHostReverseProxy(cfg.Neos.URL)

	singleJoiningSlash := func(a, b string) string {
		aslash := strings.HasSuffix(a, "/")
		bslash := strings.HasPrefix(b, "/")
		switch {
		case aslash && bslash:
			return a + b[1:]
		case !aslash && !bslash:
			return a + "/" + b
		}
		return a + b
	}

	reverseProxy.Director = func(req *http.Request) {
		// reset and rewrite request headers
		headers := http.Header{}
		headers.Set("X-Forwarded-Host", req.Host)
		headers.Set("X-Origin-Host", cfg.Neos.URL.Host)
		req.Header = headers
		req.URL.Scheme = cfg.Neos.URL.Scheme
		req.URL.Host = cfg.Neos.URL.Host
		req.Host = cfg.Neos.URL.Host

		// strip prefix
		reqURI := strings.TrimPrefix(req.URL.Path, cfg.Proxy.BasePath)
		proxyPath := singleJoiningSlash(cfg.Neos.URL.Path, reqURI)
		if strings.HasSuffix(proxyPath, "/") && len(proxyPath) > 1 {
			proxyPath = proxyPath[:len(proxyPath)-1]
		}
		req.URL.Path = proxyPath
	}

	p := &Proxy{
		log:             logging.GetDefaultLogEntry(),
		config:          cfg,
		workspaceCaches: make(map[string]*cache.Cache, len(cfg.Neos.Workspaces)),

		router:       mux.NewRouter(),
		proxyHandler: reverseProxy,

		status: &model.Status{
			Workspaces:      cfg.Neos.Workspaces,
			ProviderReports: map[string]model.Report{},
			ConsumerReports: map[string]model.Report{},
		},
		broker:             notifier.NewBroker(),
		servedStatsChan:    make(chan bool),
		servedStatsCounter: uint(0),
	}

	go func() {
		tick := time.Tick(time.Minute * 1)
		for {
			select {
			case <-tick:
				p.log.WithField("requests", p.servedStatsCounter).Info("requests served in the last 60 seconds")
				p.servedStatsCounter = uint(0)
			case <-p.servedStatsChan:
				p.servedStatsCounter++
			}
		}
	}()

	// content cache for html from neos
	p.contentCache = content_cache.New(cacheLifetime, contentStore, contentLoader, p.broker, p.log)

	// sitemap / site structure cache for content servers
	for _, workspace := range cfg.Neos.Workspaces {
		p.workspaceCaches[workspace] = cache.New(p.broker, workspace, cfg)
	}

	// setup routes
	p.setupRoutes()

	// append oberservers
	for _, observer := range cfg.Observer {
		if observer.Webhook == nil {
			continue
		}

		l := logging.GetDefaultLogEntry().WithField("url", observer.Webhook.URL).WithField("name", observer.Webhook.Name)
		l.Debug("register sitemap observer")

		n := notifier.NewContentServerNotifier(observer.Webhook.Name, observer.Webhook.URL, observer.Webhook.Token, observer.Webhook.VerifyTLS)

		for workspace, subscribers := range cfg.Subscriptions {
			for _, subscriber := range subscribers {
				if subscriber == n.GetName() {
					p.broker.RegisterSitemapObserver(workspace, n)
					l.WithField("workspace", workspace).Debug("notifier/observer registered at workspace")
				}
			}
		}

	}
	return p
}
