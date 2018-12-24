package store

// Store for cached objects
type Store interface {
	Cache() CacheStore
}
