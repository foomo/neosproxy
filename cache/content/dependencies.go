package content

import "sync"

//-----------------------------------------------------------------------------
// ~ CACHE DEPENDENCIES for all dimensions in all workspaces
//-----------------------------------------------------------------------------

type cacheDependencies struct {
	dependencies map[string]*cacheDependency
}

func NewCacheDependencies() *cacheDependencies {
	return &cacheDependencies{
		dependencies: make(map[string]*cacheDependency, 4),
	}
}

func (c *cacheDependencies) getHash(dimension, workspace string) string {
	return workspace + "_" + dimension
}

func (c *cacheDependencies) Get(id, dimension, workspace string) []string {
	hash := c.getHash(dimension, workspace)
	if cache, ok := c.dependencies[hash]; ok {
		return cache.Get(id)
	}
	return nil
}

func (c *cacheDependencies) Set(sourceID, targetID, dimension, workspace string) {
	hash := c.getHash(dimension, workspace)
	if _, ok := c.dependencies[hash]; !ok {
		c.dependencies[hash] = &cacheDependency{}
	}
	cache := c.dependencies[hash]
	cache.Set(sourceID, targetID)
	return
}

//-----------------------------------------------------------------------------
// ~ CACHE DEPENDENCY for a dimension in a workspace
//-----------------------------------------------------------------------------

type cacheDependency struct {
	lock         sync.RWMutex
	dependencies map[string][]string
}

func (c *cacheDependency) Get(id string) []string {
	c.lock.RLock()
	if dependencies, ok := c.dependencies[id]; ok {
		c.lock.RUnlock()
		return dependencies
	}
	c.lock.RUnlock()
	return nil
}

func (c *cacheDependency) Set(sourceID, targetID string) {
	c.lock.Lock()
	if c.dependencies == nil || len(c.dependencies) == 0 {
		c.dependencies = make(map[string][]string)
	}
	if _, ok := c.dependencies[targetID]; !ok {
		c.dependencies[targetID] = []string{}
	}
	// match := false
	// for _, id := range c.dependencies[targetID] {
	// 	if id == sourceID {
	// 		match = true
	// 		break
	// 	}
	// }
	// if !match {
	// 	c.dependencies[targetID] = append(c.dependencies[targetID], sourceID)
	// }
	c.dependencies[targetID] = append(c.dependencies[targetID], sourceID)
	c.lock.Unlock()
	return
}
