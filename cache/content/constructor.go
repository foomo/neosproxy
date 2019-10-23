package content

import (
	"container/list"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
	"github.com/foomo/neosproxy/logging"
	"golang.org/x/sync/singleflight"
)

// New will return a newly created content cache
func New(cacheLifetime time.Duration, store store.CacheStore, loader cms.ContentLoader, observer Observer, log logging.Entry) *Cache {
	c := &Cache{
		observer: observer,
		loader:   loader,
		store:    store,

		cacheDependencies: NewCacheDependencies(),

		invalidationRequestGroup: &singleflight.Group{},
		invalidationChannel:      make(chan InvalidationRequest, 10000),
		invalidationRetryChannel: make(chan InvalidationRequest),
		retryQueue:               &list.List{},
		lifetime:                 cacheLifetime,
		log:                      log,
	}

	// load cache dependencies
	cacheDependencies, errCacheDependencies := c.store.GetAllCacheDependencies()
	if errCacheDependencies != nil {
		c.log.WithError(errCacheDependencies).Error("unable to init cache dependencies")
	}

	// update cache dependencies
	for _, obj := range cacheDependencies {
		for _, targetID := range obj.Dependencies {
			c.cacheDependencies.Set(obj.ID, targetID, obj.Dimension, obj.Workspace)
		}
	}

	// initialize invalidation workers
	for w := 1; w <= 15; w++ {
		go c.invalidationWorker(w)
	}

	// initialize retry worker
	c.runRetryWorker()

	return c
}
