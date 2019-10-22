package proxy

import (
	"net/http/httputil"
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

	p := &Proxy{
		log:             logging.GetDefaultLogEntry(),
		maintenance:     false,
		config:          cfg,
		workspaceCaches: make(map[string]*cache.Cache, len(cfg.Neos.Workspaces)),

		router:       mux.NewRouter(),
		proxyHandler: httputil.NewSingleHostReverseProxy(cfg.Neos.URL),

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
