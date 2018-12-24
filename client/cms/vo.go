package cms

import (
	"net/http"
	"net/url"
)

// Client for a NEOS cms
type Client struct {
	// HTTP client used to communicate with NEOS CMS
	client   *http.Client
	Endpoint *url.URL

	CMS Service
}

// Content returned on getContent
type Content struct {
	HTML       string `json:"html"`
	ValidUntil int64  `json:"validUntil"`
}
