package neosproxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/foomo/neosproxy/notifier/slack"
)

const NotifierUpdateEvent = "updated"

// NotifyOnUpdate will notify all hooks on given workspace update event
func (p *Proxy) NotifyOnUpdate(workspace string, user string) {
	for _, hook := range p.Config.Callbacks.NotifyOnUpdateHooks {
		if hook.Workspace == workspace {

			// channel
			channel, ok := p.Config.Channels[hook.Channel]
			if !ok {
				log.Println(fmt.Sprintf("channel %s not configured", hook.Channel))
				continue
			}

			// notify
			go p.notify(channel, workspace, user, NotifierUpdateEvent)
		}
	}
}

// notify notifies callbacks for the given event
func (p *Proxy) notify(channel *Channel, workspace string, user string, event string) bool {

	// user
	if user == "" {
		user = "anonymous"
	}

	// logging
	log.Println(fmt.Sprintf("notifying channel %s on %s event for workspace %s, requested by %s", channel.GetChannelType(), event, workspace, user))

	switch channel.GetChannelType() {
	case ChannelTypeFoomo:
		return p.notifyDefault(channel, workspace, event, nil, nil, 0)
	case ChannelTypeSlack:
		client := slack.NewClient(channel.URL.String())
		field := slack.Field{
			Title: "NEOS Content Export",
			Value: strings.ToUpper(user) + " published a NEOS content export to " + strings.ToUpper(workspace),
			Short: false,
		}
		fields := []slack.Field{field}
		attachment := &slack.Attachment{
			Fallback: strings.ToUpper(user) + " published a NEOS content export to " + strings.ToUpper(workspace),
			Color:    "warning",
			Fields:   fields,
		}
		attachments := []*slack.Attachment{attachment}
		msg := &slack.Message{
			Channel:     channel.SlackChannel,
			IconEmoji:   ":ghost:",
			UserName:    "neosproxy",
			Attachments: attachments,
		}
		err := client.SendMessage(msg)
		return err == nil
	default:
		return p.notifyDefault(channel, workspace, event, nil, nil, 0)
	}
}

func (p *Proxy) notifyDefault(channel *Channel, workspace string, event string, payload []byte, httpClient *http.Client, redirectCount int) bool {

	// redirect counter
	if redirectCount >= 10 {
		log.Println("request failed with too many redirects")
		return false
	}

	// custom error to know if a redirect happened
	var RedirectAttemptedError = errors.New("redirect")

	// client
	if httpClient == nil {
		httpClient = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return RedirectAttemptedError
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !channel.VerifyTLS,
				},
			},
		}
	}

	// payload
	if len(payload) == 0 {
		payload, _ = json.Marshal(map[string]string{
			"type":      event,
			"workspace": workspace,
		})
	}

	// prepare request
	request, err := http.NewRequest(channel.Method, channel.URL.String(), bytes.NewBuffer(payload))
	if err != nil {
		log.Println(fmt.Sprintf("Failed to create callback request! Got error: %s", err.Error()))
		return false
	}

	// add header
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("key", channel.APIKey)

	// call
	response, err := httpClient.Do(request)

	// redirect handler
	if urlError, ok := err.(*url.Error); ok && urlError.Err == RedirectAttemptedError {
		location := response.Header.Get("location")
		channelURL, urlError := url.Parse(location)
		if urlError != nil {
			log.Println("failed to notify webhook due to invalid redirect location:", location)
			return false
		}
		channel.URL = channelURL
		redirectCount++
		return p.notifyDefault(channel, workspace, event, payload, httpClient, redirectCount)
	}

	// request error handling
	if err != nil {
		log.Println(fmt.Sprintf("failed to notify a webhook! Got error: %s", err.Error()))
		return false
	}

	log.Println(fmt.Sprintf("notified webhook with response code: %d", response.StatusCode))
	return response.StatusCode == http.StatusOK
}
