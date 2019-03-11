package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/foomo/neosproxy/config"
	"github.com/foomo/neosproxy/logging"
)

// New will return a newly created cache object
func New(broker Broker, workspace string, cfg *config.Config) *Cache {

	cacheDir := filepath.Join(cfg.Cache.Directory, "cse")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		logging.GetDefaultLogEntry().WithError(err).Fatal("failed creating cache directory")
	}

	c := &Cache{
		Workspace:           workspace,
		invalidationChannel: make(chan time.Time, 1),

		broker:   broker,
		file:     fmt.Sprintf("%s/contentserver-export-%s.json", cacheDir, workspace),
		FileLock: sync.RWMutex{},

		neos:   cfg.Neos,
		config: cfg.Cache,
	}
	go c.scheduleInvalidation()
	return c
}
