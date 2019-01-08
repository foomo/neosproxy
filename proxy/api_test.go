package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAcceptHeader(t *testing.T) {

	var accept mime

	accept = parseAcceptHeader("text/plain,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	assert.Equal(t, string(mimeTextPlain), string(accept))

	accept = parseAcceptHeader("application/xhtml+xml,application/json;q=0.9,image/webp,image/apng,*/*;q=0.8")
	assert.Equal(t, string(mimeApplicationJSON), string(accept))

	accept = parseAcceptHeader("application/xhtml+xml")
	assert.Equal(t, string(mimeApplicationJSON), string(accept))

}
