package memory

import (
	"sync"

	"github.com/foomo/neosproxy/cache/content"
	"github.com/foomo/neosproxy/cache/content/store"
)

//------------------------------------------------------------------
// ~ TYPES
//------------------------------------------------------------------

type memoryCacheStore struct {
	items map[string]store.CacheItem
	lock  *sync.RWMutex
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewCacheStore creates a new in-memory cache store
func NewCacheStore() store.CacheStore {
	s := &memoryCacheStore{
		items: map[string]store.CacheItem{},
		lock:  &sync.RWMutex{},
	}
	return s
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (s *memoryCacheStore) Upsert(cache store.CacheItem) (e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items[cache.Hash] = cache
	return
}

func (s *memoryCacheStore) Get(hash string) (cache store.CacheItem, e error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	cache, ok := s.items[hash]
	if !ok {
		e = content.ErrorNotFound
	}
	return
}

func (s *memoryCacheStore) GetAll() (caches []store.CacheItem, e error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	caches = make([]store.CacheItem, len(s.items))
	i := 0
	for _, cache := range s.items {
		caches[i] = cache
		i++
	}

	return
}

func (s *memoryCacheStore) Count() (int, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.items), nil
}

func (s *memoryCacheStore) Remove(hash string) (e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.items, hash)
	return
}

func (s *memoryCacheStore) RemoveAll() (e error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = map[string]store.CacheItem{}
	return
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------
