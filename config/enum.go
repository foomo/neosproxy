package config

type ObserverType string

const (
	ObserverTypeFoomo   ObserverType = "foomo"
	ObserverTypeSlack   ObserverType = "slack"
	ObserverTypeWebhook ObserverType = "webhook"
)
