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

// EventType indicates the payload carried by an Event.
type EventType int

const (
	EventTypeResult EventType = iota
	EventTypeProgress
	EventTypeError
)

// Event is the unified message sent over the scanner output channel.
type Event struct {
	Type     EventType
	Result   ResultEvent
	Progress ProgressEvent
	Err      error
}

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
