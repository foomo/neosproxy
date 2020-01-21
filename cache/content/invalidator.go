package content

import (
	"context"
	"strings"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/sirupsen/logrus"
)

// RemoveAll will reset whole cache by dropping all items
func (c *Cache) RemoveAll() (err error) {
	return c.store.RemoveAll()
}

// Invalidate creates an invalidation job and adds it to the queue
// serveral workers will take care of job execution
func (c *Cache) Invalidate(id, dimension string) {
	req := InvalidationRequest{
		CreatedAt:        time.Now(),
		ID:               id,
		Dimension:        dimension,
		ExecutionCounter: 0,
	}

	logger := c.log.WithFields(logrus.Fields{
		"id":                          id,
		"dimension":                   dimension,
		"lenInvalidationChannel":      len(c.invalidationChannel),
		"capInvalidationChannel":      cap(c.invalidationChannel),
		"lenInvalidationRetryChannel": len(c.invalidationRetryChannel),
		"capInvalidationRetryChannel": cap(c.invalidationRetryChannel),
	})

	select {
	case c.invalidationChannel <- req:
		logger.Info("content cache invalidation request added to invalidation queue")
		return
	default:
		logger.Info("content cache invalidation request added to retry queue")
		c.retry(req)
		return
	}
}

// Load will immediately load content from NEOS and persist it as a cache item
// no retry if it fails
func (c *Cache) Load(id, dimension string) (item store.CacheItem, err error) {

	groupName := strings.Join([]string{"invalidate", id, dimension}, "-")
	itemInterfaced, errThrottled, _ := c.invalidationRequestGroup.Do(groupName, func() (i interface{}, e error) {
		return c.invalidate(InvalidationRequest{
			CreatedAt: time.Now(),
			ID:        id,
			Dimension: dimension,
		})
	})

	if errThrottled != nil {
		err = errThrottled
		return
	}

	item = itemInterfaced.(store.CacheItem)
	return
}

// invalidate cache item, load fresh content from NEOS
func (c *Cache) invalidate(req InvalidationRequest) (item store.CacheItem, err error) {

	// timer
	start := time.Now()

	timeout := 10 * time.Second
	if req.ExecutionCounter >= 5 {
		timeout = 30 * time.Second
	}

	// context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// load item
	cmsContent, errGetContent := c.loader.GetContent(req.ID, req.Dimension, c.workspace, ctx)
	if errGetContent != nil {
		err = errGetContent
		return
	}

	// update cache dependencies
	if len(cmsContent.CacheDependencies) > 0 {
		for _, targetID := range cmsContent.CacheDependencies {
			c.cacheDependencies.Set(req.ID, targetID, req.Dimension)
		}
	}

	// invalidate dependencies
	dependencies := c.cacheDependencies.Get(req.ID, req.Dimension)
	if len(dependencies) > 0 {
		for _, nodeID := range dependencies {
			c.Invalidate(nodeID, req.Dimension)
		}
	}

	// prepare cache item
	item = store.NewCacheItem(req.ID, req.Dimension, cmsContent.HTML, cmsContent.CacheDependencies, c.validUntil(cmsContent.ValidUntil))

	// write item to cache
	errUpsert := c.store.Upsert(item)
	if errUpsert != nil {
		err = errUpsert
		return
	}

	// logging
	c.log.WithFields(logrus.Fields{
		"id":        req.ID,
		"dimension": req.Dimension,
		"retry":     req.ExecutionCounter,
		"createdAt": req.CreatedAt,
		"waitTime":  time.Since(req.CreatedAt).Seconds(),
	}).WithDuration(start).Info("content cache invalidated")

	// notify observer
	c.observer.Notify(InvalidationResponse{
		Item:     item,
		Duration: time.Since(start),
	})

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
