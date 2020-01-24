package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/foomo/neosproxy/utils"
)

var _ Notifier = &Webhook{}

type Webhook struct {
	name      string
	token     string
	endpoint  *url.URL
	client    *http.Client
	workspace string
}

// NewContentServerNotifier will create a new contentserver update notifier
func NewWebhookNotifier(name string, endpoint *url.URL, token string, verifyTLS bool, workspace string) *Webhook {

	// client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return redirectAttemptedError
		},
		Transport: utils.GetDefaultTransport(verifyTLS),
	}

	return &Webhook{
		name:      name,
		endpoint:  endpoint,
		token:     token,
		client:    client,
		workspace: workspace,
	}
}

func (wh *Webhook) GetName() string {
	return wh.name
}

// func (p *Proxy) Notify(channel *Channel, workspace string, event string, payload []byte, httpClient *http.Client, redirectCount int) bool {
func (wh *Webhook) Notify(event NotifyEvent) error {

	// // redirect counter
	// if redirectCount >= 10 {
	// 	log.Println("request failed with too many redirects")
	// 	return false
	// }

	// payload
	payload, errPayload := json.Marshal(map[string]string{
		"type":      "updated",
		"workspace": wh.workspace,
	})
	if errPayload != nil {
		return errPayload
	}

	// prepare request
	request, errRequest := http.NewRequest(http.MethodPost, wh.endpoint.String(), bytes.NewBuffer(payload))
	if errRequest != nil {
		return errRequest
	}

	// add header
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("key", wh.token)

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
