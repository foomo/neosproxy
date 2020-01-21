package content

import (
	"container/list"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
	"github.com/foomo/neosproxy/logging"
	"golang.org/x/sync/singleflight"
)

// Cache items
type Cache struct {
	observer  Observer
	loader    cms.ContentLoader
	store     store.CacheStore
	workspace string

	invalidationRequestGroup *singleflight.Group
	invalidationChannel      chan InvalidationRequest
	invalidationRetryChannel chan InvalidationRequest
	retryQueue               *list.List

	cacheDependencies *cacheDependencies
	lifetime          time.Duration // time until an item must be re-invalidated (< 0 === never)

	log logging.Entry
}

// InvalidationRequest request VO
type InvalidationRequest struct {
	ID        string
	Dimension string

	CreatedAt        time.Time
	LastExecutedAt   time.Time
	ExecutionCounter int
}

// InvalidationResponse response VO
type InvalidationResponse struct {
	Duration time.Duration
	Item     store.CacheItem
}

// Observer must be implemented by observers which are interested in update events
type Observer interface {
	Notify(response InvalidationResponse)
}
