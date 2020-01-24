package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/foomo/neosproxy/utils"
)

var _ Notifier = &Neosproxy{}

type Neosproxy struct {
	name     string
	token    string
	endpoint *url.URL
	client   *http.Client
}

var redirectAttemptedError = errors.New("redirect")

// NewContentServerNotifier will create a new contentserver update notifier
func NewNeosproxyNotifier(name string, endpoint *url.URL, token string, verifyTLS bool) *Neosproxy {

	// client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return redirectAttemptedError
		},
		Transport: utils.GetDefaultTransport(verifyTLS),
	}

	return &Neosproxy{
		name:     name,
		endpoint: endpoint,
		token:    token,
		client:   client,
	}
}

func (wh *Neosproxy) GetName() string {
	return wh.name
}

// func (p *Proxy) Notify(channel *Channel, workspace string, event string, payload []byte, httpClient *http.Client, redirectCount int) bool {
func (wh *Neosproxy) Notify(event NotifyEvent) error {

	// // redirect counter
	// if redirectCount >= 10 {
	// 	log.Println("request failed with too many redirects")
	// 	return false
	// }

	// payload
	payload, errPayload := json.Marshal(map[string]string{
		"type": "updated",
	})
	if errPayload != nil {
		return errPayload
	}

	// prepare request
	request, errRequest := http.NewRequest(http.MethodDelete, wh.endpoint.String(), bytes.NewBuffer(payload))
	if errRequest != nil {
		return errRequest
	}

	// add header
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+wh.token)

	// call
	response, errRequest := wh.client.Do(request)

	// // redirect handler
	// if urlError, ok := err.(*url.Error); ok && urlError.Err == RedirectAttemptedError {
	// 	location := response.Header.Get("location")
	// 	channelURL, errParseURL := url.Parse(location)
	// 	if errParseURL != nil {
	// 		return errParseURL
	// 	}
	// 	channel.URL = channelURL
	// 	redirectCount++
	// 	return p.notifyDefault(channel, workspace, event, payload, httpClient, redirectCount)
	// }

	// request error handling
	if errRequest != nil {
		return errRequest
	}

	if response.StatusCode != http.StatusOK {
		return errors.New("unexpected status code")
	}

	return nil
}
