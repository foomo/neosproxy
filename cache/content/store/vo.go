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

	created    time.Time
	validUntil time.Time

	HTML         string
	Etag         string // hashed fingerprint of html content
	Dependencies []string
}

type CacheDependencies struct {
	ID           string
	Dimension    string
	Dependencies []string
}

// NewCacheItem will create a new cache item
func NewCacheItem(id string, dimension string, html string, dependencies []string, validUntil time.Time) CacheItem {
	return CacheItem{
		Hash:         GetHash(id, dimension),
		ID:           id,
		Dimension:    dimension,
		created:      time.Now(),
		validUntil:   validUntil,
		HTML:         html,
		Etag:         generateFingerprint(html),
		Dependencies: dependencies,
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
func GetHash(id, dimension string) string {
	return strings.Join([]string{dimension, id}, "_")
}

func generateFingerprint(data string) string {
	sha := sha256.New()
	sha.Write([]byte(data))
	return hex.EncodeToString(sha.Sum(nil))
}
