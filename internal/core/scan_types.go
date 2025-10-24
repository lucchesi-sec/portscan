package core

import "time"

// ScanState represents the state of a probed port.
type ScanState string

const (
	StateOpen     ScanState = "open"
	StateClosed   ScanState = "closed"
	StateFiltered ScanState = "filtered"
)

// ResultEvent captures the outcome of a single port probe.
type ResultEvent struct {
	Host     string
	Port     uint16
	State    ScanState
	Banner   string
	Duration time.Duration
	Protocol string // "tcp" or "udp"
}

// ProgressEvent reports high-level scanning progress.
type ProgressEvent struct {
	Total     int
	Completed int
	Rate      float64
}

// EventKind identifies the type of event
type EventKind string

const (
	EventKindResult   EventKind = "result"
	EventKindProgress EventKind = "progress"
	EventKindError    EventKind = "error"
)

// Event is a typed envelope for all scanner events
type Event struct {
	Kind     EventKind
	Result   *ResultEvent
	Progress *ProgressEvent
	Error    error
}

// Helper constructors
func NewResultEvent(r ResultEvent) Event {
	return Event{Kind: EventKindResult, Result: &r}
}

func NewProgressEvent(p ProgressEvent) Event {
	return Event{Kind: EventKindProgress, Progress: &p}
}

func NewErrorEvent(err error) Event {
	return Event{Kind: EventKindError, Error: err}
}

// EventType is deprecated, use EventKind instead
type EventType int

const (
	EventTypeResult EventType = iota
	EventTypeProgress
	EventTypeError
)

// ScanTarget represents a host with a set of ports to scan.
type ScanTarget struct {
	Host  string
	Ports []uint16
}

// scanJob represents a single host/port pair fed to workers.
type scanJob struct {
	host string
	port uint16
}

func totalPortCount(targets []ScanTarget) int {
	total := 0
	for _, t := range targets {
		total += len(t.Ports)
	}
	return total
}
