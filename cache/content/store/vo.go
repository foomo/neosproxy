package store

import (
	"strings"
	"time"
)

var ValidUntilForever = time.Unix(0, 0)

// CacheItem for content caching
type CacheItem struct {
	Hash string

	ID        string
	Dimension string
	Workspace string

	created    time.Time
	validUntil time.Time

	HTML string
}

// NewCacheItem will create a new cache item
func NewCacheItem(id string, dimension string, workspace string, html string, validUntil time.Time) CacheItem {
	return CacheItem{
		Hash:       GetHash(id, dimension, workspace),
		ID:         id,
		Dimension:  dimension,
		Workspace:  workspace,
		created:    time.Now(),
		validUntil: validUntil,
		HTML:       html,
	}
}

// GetHash will return a cache item hash
func GetHash(id, dimension, workspace string) string {
	return strings.Join([]string{workspace, dimension, id}, "_")
}
