package cache

import (
	"sync"
	"time"

	"github.com/foomo/neosproxy/config"
)

// Cache workspace items
type Cache struct {
	Workspace           string
	invalidationChannel chan time.Time

	file     string
	FileLock sync.RWMutex

	config config.Cache
	neos   config.Neos

	broker Broker
}

// Broker to handle content structure changes
type Broker interface {
	NotifyOnSitemapChange(workspace string)
}
