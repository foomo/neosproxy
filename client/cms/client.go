package cms

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/foomo/neosproxy/utils"
	"github.com/pkg/errors"
)

//------------------------------------------------------------------
// ~ CONSTANTS / VARS
//------------------------------------------------------------------

const defaultContentType = "application/json"

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

// New will create a new CMS client
func New(endpoint string) (*Client, error) {

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse the url from 'endpoint'")
	}

	transport, transportErr := utils.GetDefaultTransportFor("CMS")
	if transportErr != nil {
		return nil, transportErr
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 5,
		Transport: transport,
	}

	c := &Client{
		client:   httpClient,
		Endpoint: endpointURL,
	}

	c.CMS = NewCMSService(c)

	return c, nil
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

// NewGetRequest is made whenever we need to get data
func (c *Client) NewGetRequest(path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a new url from path")
	}

	u := c.Endpoint.ResolveReference(rel)

	buf := &bytes.Buffer{}
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, errors.Wrap(err, "could not serialize request body")
		}
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), buf)
	req.Header.Add("Content-Type", defaultContentType)
	return req, err
}

//Do will send an API request and returns the API response (decodes and serializes)
func (c *Client) Do(req *http.Request, v interface{}) *ClientError {

	resp, err := c.client.Do(req)
	if err != nil {
		return CreateClientError(errors.Wrap(err, "could not create a new request"), req, resp, nil)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CreateClientError(errors.Wrap(err, "could not read response body"), req, resp, data)
	}
	if isHTPPResponse2xx(resp) && v != nil && len(data) > 0 {
		switch v.(type) {
		case *[]byte:
			*v.(*[]byte) = data
		default:
			if err := json.Unmarshal(data, v); err != nil {
				return CreateClientError(errors.Wrap(err, "could not unmarshal request for the specified interface"), req, resp, data)
			}
		}
	} else if !isHTPPResponse2xx(resp) {
		return CreateClientError(errors.New("cms client error"), req, resp, data)
	}
	return nil
}

//------------------------------------------------------------------
// ~ PRIVATE METHODS
//------------------------------------------------------------------

func isHTPPResponse2xx(response *http.Response) bool {
	return response.StatusCode >= 200 && response.StatusCode < 300
}
