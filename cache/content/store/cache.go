package store

// CacheStore is a store interface for content cache
type CacheStore interface {
	Upsert(item CacheItem) (e error)

	Get(hash string) (item CacheItem, e error)
	GetAll() (item []CacheItem, e error)

	Count() (int, error)

	Remove(hash string) (e error)
	RemoveAll() (e error)
}
