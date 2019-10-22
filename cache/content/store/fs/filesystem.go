package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/foomo/neosproxy/cache/content"
	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/logging"
)

//------------------------------------------------------------------
// ~ TYPES
//------------------------------------------------------------------

// fsCacheStore implements a Cache by caching the data directly to a cache directory.
type fsCacheStore struct {
	CacheDir string

	lock sync.Mutex
	rw   map[string]*sync.RWMutex
	l    logging.Entry

	lockEtags sync.RWMutex
	etags     map[string]string
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// NewCacheStore creates a new filesystem cache store
func NewCacheStore(cacheDir string) store.CacheStore {

	l := logging.GetDefaultLogEntry().WithField("cache", "fscache")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		l.WithError(err).Fatal("failed creating cache directory")
	}

	f := &fsCacheStore{
		CacheDir: cacheDir,

		l: l,

		lock: sync.Mutex{},
		rw:   make(map[string]*sync.RWMutex),

		lockEtags: sync.RWMutex{},
		etags:     make(map[string]string),
	}

	go f.initEtagCache()

	return f
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (f *fsCacheStore) Upsert(item store.CacheItem) (e error) {
	// key
	key := f.getItemKey(item)

	// validate etag
	if item.Etag == "" {
		item.Etag = item.GetEtag()
	}

	// serialize
	bytes, errMarshall := json.Marshal(item)
	if errMarshall != nil {
		return errMarshall
	}

	// lock
	cacheFile := f.Lock(key)

	// write to file
	errWrite := ioutil.WriteFile(cacheFile, bytes, 0644)
	if errWrite != nil {
		f.Unlock(key)
		e = errWrite
		return
	}
	f.Unlock(key)

	// update etag
	f.upsertEtag(item.Hash, item.Etag)

	return nil
}

func (f *fsCacheStore) GetAllCacheDependencies() ([]store.CacheDependencies, error) {
	start := time.Now()
	l := f.l.WithField(logging.FieldFunction, "GetAllCacheDependencies")
	files, errReadDir := ioutil.ReadDir(f.CacheDir)
	if errReadDir != nil {
		l.WithError(errReadDir).Error("failed reading cache dir")
		return nil, errReadDir
	}

	dependencies := []store.CacheDependencies{}

	counter := 0
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			index := strings.Index(filename, ".")
			if index >= 0 {
				filename = filename[0:index]
			}
			item, errGet := f.Get(filename)
			if errGet != nil {
				l.WithError(errGet).Warn("could not load cache item")
				continue
			}
			counter++
			dependencies = append(dependencies, store.CacheDependencies{
				ID:           item.ID,
				Dimension:    item.Dimension,
				Workspace:    item.Workspace,
				Dependencies: item.Dependencies,
			})
		}
	}
	l.WithField("len", counter).WithDuration(start).Debug("all cache dependencies loaded")

	return dependencies, nil
}

func (f *fsCacheStore) GetAllEtags(workspace string) (etags map[string]string) {
	f.lockEtags.RLock()
	etags = make(map[string]string)
	for hash, etag := range f.etags {
		if !strings.HasPrefix(hash, workspace) {
			continue
		}
		etags[hash] = etag
	}
	f.lockEtags.RUnlock()
	return
}

func (f *fsCacheStore) GetEtag(hash string) (etag string, e error) {
	f.lockEtags.RLock()
	if value, ok := f.etags[hash]; ok {
		etag = value
		f.lockEtags.RUnlock()
		return
	}
	f.lockEtags.RUnlock()

	item, errGet := f.Get(hash)
	if errGet != nil {
		e = errGet
		return
	}

	etag = item.GetEtag()
	f.upsertEtag(hash, etag)

	return
}

func (f *fsCacheStore) Get(hash string) (item store.CacheItem, e error) {
	key := f.getKey(hash)
	cacheFile, _ := f.RLock(key)

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		f.RUnlock(key)
		e = content.ErrorNotFound
		return
	}

	bytes, errReadFile := ioutil.ReadFile(cacheFile)
	if errReadFile != nil {
		f.RUnlock(key)
		e = errReadFile
		return
	}

	f.RUnlock(key)

	item = store.CacheItem{}
	errUnmarshall := json.Unmarshal(bytes, &item)
	if errUnmarshall != nil {
		go f.Remove(hash)
		e = errUnmarshall
		return
	}

	return
}

func (f *fsCacheStore) GetAll() (items []store.CacheItem, e error) {
	files, errReadDir := ioutil.ReadDir(f.CacheDir)
	if errReadDir != nil {
		e = errReadDir
		return
	}

	items = []store.CacheItem{}
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			index := strings.Index(filename, ".")
			if index >= 0 {
				filename = filename[0:index]
			}
			item, errGet := f.Get(filename)
			if errGet != nil {
				e = errGet
				return
			}
			items = append(items, item)
		}
	}

	return
}

func (f *fsCacheStore) Count() (int, error) {
	i := 0
	files, err := ioutil.ReadDir(f.CacheDir)
	if err != nil {
		return 0, err
	}
	for _, file := range files {
		if !file.IsDir() {
			i++
		}
	}
	return i, nil
}

func (f *fsCacheStore) Remove(hash string) (e error) {
	key := f.getKey(hash)
	cacheFile := f.Lock(key)
	defer f.Unlock(key)

	errRemove := os.Remove(cacheFile)
	if errRemove != nil {
		e = errRemove
		return
	}

	f.lockEtags.Lock()
	delete(f.etags, hash)
	f.lockEtags.Unlock()

	return nil
}

func (f *fsCacheStore) createCacheDir() error {
	return os.MkdirAll(f.CacheDir, 0755)
}

func (f *fsCacheStore) RemoveAll() (e error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	errRemoveAll := os.RemoveAll(f.CacheDir)
	if errRemoveAll != nil {
		f.l.WithError(errRemoveAll).Error("unable to remove all files from cache")
		return errRemoveAll
	}

	f.lockEtags.Lock()
	f.etags = make(map[string]string)
	f.lockEtags.Unlock()

	errCreateCache := f.createCacheDir()
	if errCreateCache != nil {
		f.l.WithError(errCreateCache).Error("unable to re-create cache directory")
		return errCreateCache
	}

	return nil
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func (f *fsCacheStore) initEtagCache() {
	start := time.Now()
	l := f.l.WithField(logging.FieldFunction, "initEtagCache")
	files, errReadDir := ioutil.ReadDir(f.CacheDir)
	if errReadDir != nil {
		l.WithError(errReadDir).Error("failed reading cache dir")
		return
	}

	l.Debug("init etag cache")
	counter := 0
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			index := strings.Index(filename, ".")
			if index >= 0 {
				filename = filename[0:index]
			}
			item, errGet := f.Get(filename)
			if errGet != nil {
				l.WithError(errGet).Warn("could not load cache item")
				continue
			}
			counter++
			f.upsertEtag(item.Hash, item.GetEtag())
		}
	}
	l.WithField("len", counter).WithDuration(start).Debug("etag cache initialized")

	return
}

func (f *fsCacheStore) upsertEtag(hash, etag string) {
	f.lockEtags.Lock()
	f.etags[hash] = etag
	f.lockEtags.Unlock()
}

func (f *fsCacheStore) getItemKey(item store.CacheItem) string {
	return f.getKey(item.Hash)
}

func (f *fsCacheStore) getKey(hash string) string {
	return hash + ".json"
}
