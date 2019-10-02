package store

import (
	"crypto/sha256"
	"encoding/hex"
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
	Etag string // hashed fingerprint of html content
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
		Etag:       generateFingerprint(html),
	}
}

// GetEtag returns an etag
func (item *CacheItem) GetEtag() string {
	if item.Etag != "" {
		return item.Etag
	}
	return generateFingerprint(item.HTML)
}

// GetHash will return a cache item hash
func GetHash(id, dimension, workspace string) string {
	return strings.Join([]string{workspace, dimension, id}, "_")
}

func generateFingerprint(data string) string {
	sha := sha256.New()
	sha.Write([]byte(data))
	return hex.EncodeToString(sha.Sum(nil))
}
