package memory

import (
	"testing"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	s := NewCacheStore()

	id := "1234"
	dimension := "de"
	workspace := "stage"
	validUntil := time.Now()

	count, countErr := s.Count()
	assert.NoError(t, countErr)
	assert.Equal(t, 0, count)

	dependencies := []string{"foo", "bar"}

	item := store.NewCacheItem(id, dimension, workspace, "<h1>Test</h1>", dependencies, validUntil)
	hash := item.Hash
	errUpsert := s.Upsert(item)
	assert.NoError(t, errUpsert)

	assert.Equal(t, hash, store.GetHash(id, dimension, workspace))
	assert.NotEmpty(t, hash)
	assert.NotEmpty(t, store.GetHash(id, dimension, workspace))

	count, countErr = s.Count()
	assert.NoError(t, countErr)
	assert.Equal(t, 1, count)

	itemCached, errGet := s.Get(hash)
	assert.NoError(t, errGet)
	assert.NotNil(t, itemCached)

	assert.Equal(t, 2, len(itemCached.Dependencies))

	errRemoveAll := s.RemoveAll()
	assert.NoError(t, errRemoveAll)

	countAfterRemove, errCountAfterRemove := s.Count()
	assert.NoError(t, errCountAfterRemove)
	assert.Equal(t, 0, countAfterRemove)
}
