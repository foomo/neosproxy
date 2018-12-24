package config

import "net/url"

//-----------------------------------------------------------------------------
// ~ Private methods
//-----------------------------------------------------------------------------

func newObserverWebhook(o configFileObserver) (observer *ObserverWebhook, err error) {
	observerURL, errObserverURLParser := url.Parse(o.URL)
	if errObserverURLParser != nil {
		err = errObserverURLParser
		return
	}
	observer = &ObserverWebhook{
		Type:      ObserverTypeWebhook,
		Name:      o.Name,
		URL:       observerURL,
		Token:     o.Token,
		VerifyTLS: o.VerifyTLS,
	}
	return
}

func newObserverFoomo(o configFileObserver) (observer *ObserverFoomo, err error) {
	observerURL, errObserverURLParser := url.Parse(o.URL)
	if errObserverURLParser != nil {
		err = errObserverURLParser
		return
	}
	observer = &ObserverFoomo{
		Type:      ObserverTypeFoomo,
		Name:      o.Name,
		URL:       observerURL,
		VerifyTLS: o.VerifyTLS,
	}
	return
}

func newObserverSlack(o configFileObserver) (observer *ObserverSlack, err error) {
	observerURL, errObserverURLParser := url.Parse(o.URL)
	if errObserverURLParser != nil {
		err = errObserverURLParser
		return
	}
	observer = &ObserverSlack{
		Type:    ObserverTypeSlack,
		Name:    o.Name,
		URL:     observerURL,
		Channel: o.Channel,
	}
	return
}
