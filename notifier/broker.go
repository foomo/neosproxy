package notifier

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/foomo/neosproxy/cache"
	"github.com/foomo/neosproxy/logging"

	content_cache "github.com/foomo/neosproxy/cache/content"
)

var _ content_cache.Observer = &Broker{}
var _ cache.Broker = &Broker{}

// Broker to notify observers
type Broker struct {
	contentLock      *sync.RWMutex
	contentObservers map[string][]Notifier

	sitemapLock      *sync.RWMutex
	sitemapObservers map[string][]Notifier
}

// NewBroker will create a new message broker to handle cache invalidation notifications
func NewBroker() *Broker {
	return &Broker{
		contentLock:      &sync.RWMutex{},
		contentObservers: map[string][]Notifier{},

		sitemapLock:      &sync.RWMutex{},
		sitemapObservers: map[string][]Notifier{},
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
func (b *Broker) NotifyOnSitemapChange(workspace string) {
	b.sitemapLock.RLock()
	defer b.sitemapLock.RUnlock()

	if observers, ok := b.sitemapObservers[workspace]; ok {

		event := NotifyEvent{
			EventType: EventTypeSitemapUpdate,
			Payload:   workspace,
		}

		for _, observer := range observers {
			logging.GetDefaultLogEntry().WithFields(logrus.Fields{
				"name":      observer.GetName(),
				"workspace": workspace,
			}).Debug("broker: NotifyOnSitemapChange")

			go observer.Notify(event)
		}
	}
}

func (b *Broker) RegisterContentObserver(workspace string, observer Notifier) {
	b.contentLock.Lock()
	defer b.contentLock.Unlock()

	observers, ok := b.contentObservers[workspace]
	if !ok {
		observers = []Notifier{}
	}

	observers = append(observers, observer)
	b.contentObservers[workspace] = observers

}

func (b *Broker) RegisterSitemapObserver(workspace string, observer Notifier) {
	b.sitemapLock.Lock()
	defer b.sitemapLock.Unlock()

	observers, ok := b.sitemapObservers[workspace]
	if !ok {
		observers = []Notifier{}
	}

	observers = append(observers, observer)
	b.sitemapObservers[workspace] = observers
}
