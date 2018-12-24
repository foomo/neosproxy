package content

import (
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
)

// RemoveAll will reset whole cache by dropping all items
func (c *Cache) RemoveAll() (err error) {
	return c.store.RemoveAll()
}

// Invalidate cache item
func (c *Cache) Invalidate(id, dimension, workspace string) (item store.CacheItem, err error) {

	// timer
	start := time.Now()

	// load item
	cmsContent, errGetContent := c.loader.GetContent(id, dimension, workspace)
	if errGetContent != nil {
		err = errGetContent
		return
	}

	// prepare cache item
	item = store.NewCacheItem(id, dimension, workspace, cmsContent.HTML, c.validUntil(cmsContent.ValidUntil))

	// write item to cache
	errUpsert := c.store.Upsert(item)
	if errUpsert != nil {
		err = errUpsert
		return
	}

	// notify observer
	c.observer.Notify(InvalidationResponse{
		Item:     item,
		Duration: time.Since(start),
	})

	// done
	return
}

func (c *Cache) validUntil(validUntil int64) time.Time {

	now := time.Now()
	if validUntil > 0 && validUntil > now.Unix() {
		return time.Unix(validUntil, 0)
	}

	if c.lifetime <= 0 {
		return store.ValidUntilForever
	}

	return now.Add(c.lifetime)
}
