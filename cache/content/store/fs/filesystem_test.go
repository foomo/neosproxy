package fs

import (
	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewCacheStore(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	hash := "derp"

	s := NewCacheStore(dir)
	item := store.CacheItem{
		Hash:      hash,
		ID:        "123",
		Dimension: "de",
		Workspace: "live",
		HTML:      "<html></html>",
	}
	err = s.Upsert(item)
	assert.NoError(t, err)

	cachedItem, err := s.Get(hash)
	assert.NoError(t, err)
	assert.Equal(t, item, cachedItem)

}
