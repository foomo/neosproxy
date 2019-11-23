package proxy

import (
	"net/http/httputil"

	"github.com/foomo/neosproxy/cache"
	"github.com/foomo/neosproxy/config"
	"github.com/foomo/neosproxy/logging"
	"github.com/foomo/neosproxy/model"
	"github.com/foomo/neosproxy/notifier"
	"github.com/gorilla/mux"

	content_cache "github.com/foomo/neosproxy/cache/content"
)

// Proxy struct definition
type Proxy struct {
	log             logging.Entry
	basicAuth       []basicAuth
	config          *config.Config
	workspaceCaches map[string]*cache.Cache

	router       *mux.Router
	proxyHandler *httputil.ReverseProxy
	contentCache *content_cache.Cache

	status *model.Status
	broker *notifier.Broker

	servedStatsChan    chan bool
	servedStatsCounter uint // served requests per minute
}

type basicAuth struct {
	user     string
	password string
}
