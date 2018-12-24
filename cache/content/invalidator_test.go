package content

import (
	"testing"
	"time"

	"github.com/foomo/neosproxy/cache/content/store"

	"github.com/stretchr/testify/assert"
)

func TestValidUntil(t *testing.T) {

	var now time.Time
	var valid time.Time
	var c *Cache

	// cache forever
	c = &Cache{lifetime: 0}
	valid = c.validUntil(0)
	assert.Equal(t, store.ValidUntilForever, valid)

	// use explicit cache date, it's valid
	now = time.Now().Add(time.Minute * 5)
	c = &Cache{lifetime: 0}
	valid = c.validUntil(now.Unix())
	assert.Equal(t, now.Unix(), valid.Unix())

	// cache forever, explicit cache date is in the past
	now = time.Now().Add(time.Minute * -5)
	c = &Cache{lifetime: 0}
	valid = c.validUntil(now.Unix())
	assert.Equal(t, store.ValidUntilForever, valid)

	// default cache duration, explicit cache date is in the past
	now = time.Now().Add(time.Minute * -5)
	c = &Cache{lifetime: time.Minute * 10}
	valid = c.validUntil(now.Unix())
	assert.True(t, time.Now().Unix() <= valid.Unix())
	assert.True(t, time.Now().Add(time.Minute*10).Unix() >= valid.Unix())

	// default cache duration, explicit cache not set
	c = &Cache{lifetime: time.Minute * 10}
	valid = c.validUntil(0)
	assert.True(t, time.Now().Unix() <= valid.Unix())
	assert.True(t, time.Now().Add(time.Minute*10).Unix() >= valid.Unix())
}
