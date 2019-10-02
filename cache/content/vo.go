package content

import (
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/client/cms"
	"golang.org/x/sync/singleflight"
)

// Cache workspace items
type Cache struct {
	observer                 Observer
	loader                   cms.ContentLoader
	store                    store.CacheStore
	invalidationChannel      chan InvalidationRequest
	invalidationRequestGroup *singleflight.Group

	lifetime time.Duration // time until an item must be re-invalidated (< 0 === never)
}

// InvalidationRequest request VO
type InvalidationRequest struct {
	CreatedAt time.Time
	ID        string
	Dimension string
	Workspace string
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
