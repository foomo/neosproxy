package content

import (
	"strings"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
)

// RemoveAll will reset whole cache by dropping all items
func (c *Cache) RemoveAll() (err error) {
	return c.store.RemoveAll()
}

func (c *Cache) invalidator(id, dimension, workspace string) (item store.CacheItem, err error) {

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

	return
}

// Invalidate cache item
func (c *Cache) Invalidate(id, dimension, workspace string) (item store.CacheItem, err error) {

	groupName := strings.Join([]string{"invalidate", id, dimension, workspace}, "-")
	itemInterfaced, errThrottled, _ := c.invalidationRequestGroup.Do(groupName, func() (i interface{}, e error) {
		return c.invalidator(id, dimension, workspace)
	})

	if errThrottled != nil {
		err = errThrottled
		return
	}

	item = itemInterfaced.(store.CacheItem)
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
