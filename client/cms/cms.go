package cms

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/foomo/neosproxy/cache/content/store"
	"github.com/foomo/neosproxy/logging"
	"github.com/sirupsen/logrus"
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
func (s *cmsService) GetContent(id string, dimension string, workspace string, ctx context.Context) (content Content, e error) {

	l := s.logger.WithFields(logrus.Fields{
		logging.FieldFunction: "GetContent",
		"id":                  id,
		"dimension":           dimension,
		"workspace":           workspace,
	})
	var clientErr error

	path := fmt.Sprintf(pathContent, dimension, id, workspace)
	req, reqErr := s.client.NewGetRequest(path, nil)
	if reqErr != nil {
		l.WithError(reqErr).Error("unable to create cms get content request")
		e = ErrorRequest
		return
	}

	content = Content{}
	e, clientErr = s.convertClientErr(s.client.Do(req, ctx, &content), ctx)
	if clientErr != nil {
		l.WithError(clientErr).Error("unable to load html content from cms")
		return
	}

	return
}

//-----------------------------------------------------------------------------
// ~ PRIVATE METHODS
//-----------------------------------------------------------------------------

func (s *cmsService) convertClientErr(clientErr *ClientError, ctx context.Context) (error, error) {
	if clientErr == nil {
		return nil, nil
	}

	if ctx != nil && ctx.Err() != nil && ctx.Err() == context.DeadlineExceeded {
		return ErrorResponseTimeout, clientErr.Error
	}
	if ctx != nil && ctx.Err() != nil && ctx.Err() == context.Canceled {
		return ErrorRequest, clientErr.Error
	}

	switch clientErr.Response.StatusCode {
	case http.StatusServiceUnavailable:
		return ErrorMaintenance, clientErr.Error
	case http.StatusNotFound:
		return ErrorNotFound, clientErr.Error
	case http.StatusBadRequest:
		return ErrorBadRequest, clientErr.Error
	case http.StatusInternalServerError:
		return ErrorInternalServerError, clientErr.Error
	}

	return ErrorResponse, clientErr.Error
}
