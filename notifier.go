package neosproxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// NotifyOnUpdate will notify all hooks on given workspace update event
func (p *Proxy) NotifyOnUpdate(workspace string) {
	for _, hook := range p.Config.Callbacks.NotifyOnUpdateHooks {
		if hook.Workspace == workspace {
			go p.notify("updated", hook)
		}
	}
}

// notify notifies callbacks for the given event
func (p *Proxy) notify(event string, hook *Hook) {
	// logging
	log.Println(fmt.Sprintf("Notifying '%s' event on workspace %s: %s", event, hook.Workspace, hook.URL))

	// payload
	data, _ := json.Marshal(map[string]string{
		"type":      event,
		"workspace": hook.Workspace,
	})

	// setup http client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !hook.VerifyTLS,
			},
		},
	}

	// prepare request
	req, err := http.NewRequest(http.MethodPost, hook.URL.String(), bytes.NewBuffer(data))
	if err != nil {
		log.Println(fmt.Sprintf("Failed to create callback request! Got error: %s", err.Error()))
		return
	}

	// add header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("key", hook.APIKey)

	// send request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(fmt.Sprintf("failed to notify a webhook! Got error: %s", err.Error()))
	} else {
		log.Println(fmt.Sprintf("notified webhook with response code: %d", resp.StatusCode))
	}
}
