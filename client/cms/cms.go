package cms

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/logging"
)

//-----------------------------------------------------------------------------
// ~ Constants
//-----------------------------------------------------------------------------

const (
	pathImage      = "/contentserver/image"
	pathAsset      = "/contentserver/asset"
	pathContent    = "/contentserver/export/%s/%s?workspace=%s"
	pathRepository = "/contentserver/export"

	// WorkspaceStage constants for stage workspace
	WorkspaceStage string = "stage"
	// WorkspaceLive constants for live workspace
	WorkspaceLive string = "live"
)

//-----------------------------------------------------------------------------
// ~ CONSTANTS / VARS
//-----------------------------------------------------------------------------

// Code type checking for interface implementation
var _ Service = &cmsService{}

var _ ContentLoader = &cmsService{}

//------------------------------------------------------------------
// ~ TYPES
//------------------------------------------------------------------

type memoryCacheStore struct {
	items map[string]store.CacheItem
	lock  *sync.RWMutex

	store.CacheStore
}

//-----------------------------------------------------------------------------
// ~ TYPES
//-----------------------------------------------------------------------------

type cmsService struct {
	client *Client
	logger logging.Entry
}

//-----------------------------------------------------------------------------
// ~ CONSTRUCTOR
//-----------------------------------------------------------------------------

// NewCMSService constructed with given cms client
func NewCMSService(defaultClient *Client) Service {
	service := &cmsService{
		client: defaultClient,
		logger: logging.GetDefaultLogEntry(),
	}
	return service
}

//-----------------------------------------------------------------------------
// ~ PUBLIC METHODS
//-----------------------------------------------------------------------------

// GetContent from NEOS CMS as html string
func (s *cmsService) GetContent(id string, dimension string, workspace string) (content Content, e error) {

	l := s.logger.WithField(logging.FieldFunction, "GetContent")
	var clientErr error

	path := fmt.Sprintf(pathContent, dimension, id, workspace)
	req, reqErr := s.client.NewGetRequest(path, nil)
	if reqErr != nil {
		l.WithError(reqErr).Error("unable to create cms get content request")
		e = ErrorRequest
		return
	}

	content = Content{}
	e, clientErr = s.convertClientErr(s.client.Do(req, &content))
	if clientErr != nil {
		l.WithError(clientErr).Error("unable to load html content from cms")
		return
	}

	return
}

//-----------------------------------------------------------------------------
// ~ PRIVATE METHODS
//-----------------------------------------------------------------------------

func (s *cmsService) convertClientErr(clientErr *ClientError) (error, error) {
	if clientErr == nil {
		return nil, nil
	}

	if clientErr.Response.StatusCode == http.StatusServiceUnavailable {
		return ErrorMaintenance, clientErr.Error
	}

	return ErrorResponse, clientErr.Error
}
