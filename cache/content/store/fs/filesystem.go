package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"

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

	s := &fsCacheStore{
		CacheDir: cacheDir,
		lock:     sync.Mutex{},
		rw:       make(map[string]*sync.RWMutex),
		l:        l,
	}

	return s
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (f *fsCacheStore) Upsert(item store.CacheItem) (e error) {
	// key
	key := f.getItemKey(item)

	// serialize
	bytes, errMarshall := json.Marshal(item)
	if errMarshall != nil {
		return errMarshall
	}

	// lock
	cacheFile := f.Lock(key)
	defer f.Unlock(key)

	// write to file
	return ioutil.WriteFile(cacheFile, bytes, 0644)
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

	return os.Remove(cacheFile)
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

func (f *fsCacheStore) getItemKey(item store.CacheItem) string {
	return f.getKey(item.Hash)
}

func (f *fsCacheStore) getKey(hash string) string {
	return hash + ".json"
}
