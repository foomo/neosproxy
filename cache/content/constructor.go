package content

import (
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
)

// New will return a newly created content cache
func New(cacheLifetime time.Duration, store store.CacheStore, loader cms.ContentLoader, observer Observer) *Cache {
	c := &Cache{
		observer: observer,
		loader:   loader,
		store:    store,

		// invalidationChannel: make(chan InvalidationRequest),
		lifetime: cacheLifetime,
	}
	return c
}
