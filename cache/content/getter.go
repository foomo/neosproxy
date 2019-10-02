package content

import (
	"github.com/foomo/neosproxy/cache/content/store"
)

// Get a cache item, if it exists
func (c *Cache) Get(id, dimension, workspace string) (item store.CacheItem, err error) {

	hash := store.GetHash(id, dimension, workspace)
	cachedItem, errGet := c.store.Get(hash)
	if errGet != nil {
		err = errGet
		return
	}

	item = cachedItem
	return
}

// Len will return the number of cached items
func (c *Cache) Len() int {
	counter, errCounter := c.store.Count()
	if errCounter != nil {
		return 0
	}
	return counter
}

func (c *Cache) GetAllEtags(workspace string) (etags map[string]string) {
	return c.store.GetAllEtags(workspace)
}

func (c *Cache) GetEtag(hash string) (etag string, e error) {
	return c.store.GetEtag(hash)
}
