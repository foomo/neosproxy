package mongostore

import (
	"github.com/foomo/neosproxy/cache/content/store"
)

//------------------------------------------------------------------
// ~ INTERFACES
//------------------------------------------------------------------

type MongoStore interface {
	store.Store
}

//------------------------------------------------------------------
// ~ STRUCTS
//------------------------------------------------------------------

type mongoStores struct {
	cache store.CacheStore
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewMongoStore will create a new mongo store
func NewMongoStore(url string) (s MongoStore, e error) {

	// get cache mongo persistor
	cachePersistor, errCachePersistor := getCachePersistor(url)
	if errCachePersistor != nil {
		e = errCachePersistor
		return
	}

	// init stores
	s = &mongoStores{
		cache: NewCacheStore(cachePersistor),
	}

	return
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (s *mongoStores) Cache() store.CacheStore {
	return s.cache
}
