package model

import (
	"time"
)

type ReportStatus string
type MessageStatus string

const (
	ReportStatusValid   ReportStatus = "valid"
	ReportStatusInvalid ReportStatus = "invalid"
	ReportStatusUnknown ReportStatus = "unknown"

	MessageStatusInfo    MessageStatus = "info"
	MessageStatusWarning MessageStatus = "warning"
	MessageStatusError   MessageStatus = "error"
)

type Report struct {
	Name string
	URL  string

	Status    ReportStatus
	DateTime  time.Time
	Hash      string
	Workspace string

	Messages struct {
		Status  MessageStatus
		NodeID  string
		Message string
		Data    map[string]string
	}
}
