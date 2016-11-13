package datadog

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// Constants used to references against statsd's unexported
// `alertType`.
const (
	Error   = "error"
	Info    = "info"
	Success = "success"
)

// EventCreator is a interface for passing a statsd DataDog NewEvent
// and responsible - Google Search for creating new metric events.
type EventCreator func(title, text string) *statsd.Event

// TrackTransactionArgs is the args structs for tracking transactions
// to DataDog.
type TrackTransactionArgs struct {
	Tags            []string
	Client          *statsd.Client
	StartTime       time.Time
	AlertType       string
	EventInfo       []string
	MetricName      string
	CreateEvent     EventCreator
	CustomEventName string
}