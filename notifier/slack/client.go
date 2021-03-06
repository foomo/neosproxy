package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Message
type Message struct {
	Text        string        `json:"text"`
	Channel     string        `json:"channel,omitempty"`
	UserName    string        `json:"username,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

// Attachments
type Attachment struct {
	Fallback   string  `json:"fallback,omitempty"` // plain text summary
	Color      string  `json:"color,omitempty"`    // {good|warning|danger|hex}
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Title      string  `json:"title,omitempty"` // larger, bold text at top of attachment
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	ThumbURL   string  `json:"thumb_url,omitempty"`
	FooterIcon string  `json:"footer,omitempty"`
	Footer     string  `json:"footer_icon,omitempty"`
	Timestamp  int     `json:"ts,omitempty"` // Unix timestamp
}

// Field
type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// Add attachments to a Slack Message
func (m *Message) AddAttachment(a *Attachment) {
	m.Attachments = append(m.Attachments, a)
}

// Client
type Client struct {
	url string
}

// New Slack Incoming WebHook Client using http.DefaultClient for its Poster.
func NewClient(url string) *Client {
	return &Client{url: url}
}

// Simple text message
func (c *Client) Simple(msg string) error {
	return c.SendMessage(&Message{Text: msg})
}

// SendMessage will send a message
func (c *Client) SendMessage(msg *Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Post(c.url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Discard response body to reuse connection
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
