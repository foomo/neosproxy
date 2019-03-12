package mongostore

import (
	"github.com/foomo/neosproxy/cache/content"
	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/shop/persistence"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"

	mgo "gopkg.in/mgo.v2"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

const cacheStoreCollection = "cache"

//------------------------------------------------------------------
// ~ TYPES
//------------------------------------------------------------------

type mongoCacheStore struct {
	persistor *persistence.Persistor
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewCacheStore creates a new mongo cache store
func NewCacheStore(p *persistence.Persistor) store.CacheStore {
	s := &mongoCacheStore{persistor: p}
	s.ensureIndexes()
	return s
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (s mongoCacheStore) Upsert(cache store.CacheItem) (e error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	_, e = collection.Upsert(bson.M{"hash": cache.Hash}, cache)
	return
}

func (s mongoCacheStore) Get(hash string) (cache store.CacheItem, e error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	q := collection.Find(bson.M{"hash": hash}).Limit(1)
	errMongo := q.One(&cache)
	if errMongo != nil {
		if errMongo == mgo.ErrNotFound {
			e = content.ErrorNotFound
			return
		}
		e = errMongo
		return
	}

	return
}

func (s mongoCacheStore) GetAll() (caches []store.CacheItem, e error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	caches = make([]store.CacheItem, 0)

	q := collection.Find(bson.M{})
	e = q.All(&caches)
	return
}

func (s mongoCacheStore) Count() (int, error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	return collection.Count()
}

func (s mongoCacheStore) Remove(hash string) (e error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	e = collection.Remove(bson.M{"hash": hash})
	return
}

func (s mongoCacheStore) RemoveAll() (e error) {
	session, collection := s.persistor.GetCollection()
	defer session.Close()

	_, e = collection.RemoveAll(&bson.M{})
	return
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func (s mongoCacheStore) ensureIndexes() (e error) {

	// index definitions
	indexes := []mgo.Index{
		mgo.Index{
			Name:       "hash",
			Key:        []string{"hash"},
			Unique:     true,
			Background: true,
		},
		mgo.Index{
			Name:       "id-dimension-workspace",
			Key:        []string{"id", "dimension", "workspace"},
			Unique:     true,
			Background: true,
		},
		mgo.Index{
			Name:       "id",
			Key:        []string{"id"},
			Unique:     false,
			Background: true,
		},
		mgo.Index{
			Name:       "workspace",
			Key:        []string{"workspace"},
			Unique:     false,
			Background: true,
		},
		mgo.Index{
			Name:       "dimension",
			Key:        []string{"dimension"},
			Unique:     false,
			Background: true,
		},
	}

	// ensure indices on collection
	return s.persistor.EnsureIndexes(indexes)
}

func getCachePersistor(url string) (*persistence.Persistor, error) {
	persistor, err := getPersistor(url, cacheStoreCollection)
	if err != nil {
		return nil, errors.Wrap(err, "cache store initialization failed")
	}
	return persistor, nil
}
