package cms

import (
	"net/http"

	"github.com/pkg/errors"
)

var (
	ErrorUnknown     = errors.New("unknown error occured")
	ErrorRequest     = errors.New("request error")
	ErrorResponse    = errors.New("response error")
	ErrorMaintenance = errors.New("cms in maintenance mode")
)

type ClientError struct {
	RequestURI    string
	RequestMethod string
	Response      ClientErrorResponse
	Error         error
}

type ClientErrorResponse struct {
	Data       []byte
	StatusCode int
}

func CreateClientError(err error, request *http.Request, response *http.Response, data []byte) *ClientError {
	clientErr := &ClientError{
		RequestURI:    request.URL.RequestURI(),
		RequestMethod: request.Method,
		Error:         err,
		Response: ClientErrorResponse{
			Data: data,
		},
	}
	if response != nil {
		clientErr.Response.StatusCode = response.StatusCode
	}

	return clientErr
}
