package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/foomo/neosproxy/config"
)

// New will return a newly created cache object
func New(broker Broker, workspace string, cfg *config.Config) *Cache {
	c := &Cache{
		Workspace:           workspace,
		invalidationChannel: make(chan time.Time, 1),

		broker:   broker,
		file:     fmt.Sprintf("%s/contentserver-export-%s.json", cfg.Cache.Directory, workspace),
		FileLock: sync.RWMutex{},

		neos:   cfg.Neos,
		config: cfg.Cache,
	}
	go c.scheduleInvalidation()
	return c
}
