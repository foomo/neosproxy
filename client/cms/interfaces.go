package cms

import "context"

//-----------------------------------------------------------------------------
// ~ INTERFACES
//-----------------------------------------------------------------------------

// Service to load
type Service interface {
	GetContent(id string, dimension string, workspace string, ctx context.Context) (content Content, e error)

	// GetRepo(id string, dimension string) (html string, e error)
	// GetImage(id string, dimension string) (html string, e error)
	// GetAsset(id string, dimension string) (html string, e error)
}

// ContentLoader interface
type ContentLoader interface {
	GetContent(id, dimension, workspace string, ctx context.Context) (content Content, e error)
}
