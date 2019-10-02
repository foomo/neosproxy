package content

import (
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
	"golang.org/x/sync/singleflight"
)

// New will return a newly created content cache
func New(cacheLifetime time.Duration, store store.CacheStore, loader cms.ContentLoader, observer Observer) *Cache {
	c := &Cache{
		observer: observer,
		loader:   loader,
		store:    store,

		// invalidationChannel: make(chan InvalidationRequest),
		invalidationRequestGroup: &singleflight.Group{},
		lifetime:                 cacheLifetime,
	}
	return c
}
