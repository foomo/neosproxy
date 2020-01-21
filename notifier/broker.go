package notifier

import (
	"sync"

	"github.com/foomo/neosproxy/cache"
	"github.com/foomo/neosproxy/logging"

	content_cache "github.com/foomo/neosproxy/cache/content"
)

var _ content_cache.Observer = &Broker{}
var _ cache.Broker = &Broker{}

// Broker to notify observers
type Broker struct {
	contentLock      *sync.RWMutex
	contentObservers []Notifier

	sitemapLock      *sync.RWMutex
	sitemapObservers []Notifier
}

// NewBroker will create a new message broker to handle cache invalidation notifications
func NewBroker() *Broker {
	return &Broker{
		contentLock:      &sync.RWMutex{},
		contentObservers: []Notifier{},

		sitemapLock:      &sync.RWMutex{},
		sitemapObservers: []Notifier{},
	}
}

// Notify will be called from cache in case an item has been invalidated
func (b *Broker) Notify(response content_cache.InvalidationResponse) {
	b.contentLock.RLock()
	defer b.contentLock.RUnlock()
	// for _, observer := range b.contentObservers {
	// 	go observer.Notify(response)
	// }
}

// NotifyOnSitemapChange guess what ... will be called in case the content structure has changed
func (b *Broker) NotifyOnSitemapChange() {
	b.sitemapLock.RLock()
	defer b.sitemapLock.RUnlock()

	event := NotifyEvent{
		EventType: EventTypeSitemapUpdate,
	}

	for _, observer := range b.sitemapObservers {
		logging.GetDefaultLogEntry().WithField("name", observer.GetName()).Debug("broker: NotifyOnSitemapChange")

		go observer.Notify(event)
	}
}

func (b *Broker) RegisterContentObserver(observer Notifier) {
	b.contentLock.Lock()
	defer b.contentLock.Unlock()

	b.contentObservers = append(b.contentObservers, observer)
}

func (b *Broker) RegisterSitemapObserver(observer Notifier) {
	b.sitemapLock.Lock()
	defer b.sitemapLock.Unlock()

	b.sitemapObservers = append(b.sitemapObservers, observer)
}
