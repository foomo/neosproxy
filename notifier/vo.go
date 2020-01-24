package notifier

type EventType string

const (
	EventTypeSitemapUpdate EventType = "EventTypeSitemapUpdate"
	EventTypeContentUpdate EventType = "EventTypeContentUpdate"
)

type NotifyEvent struct {
	EventType EventType
	Payload   NotifyPayload
}

type Notifier interface {
	GetName() string
	Notify(event NotifyEvent) error
}

type NotifyPayload struct {
	ID        string
	Dimension string
}
