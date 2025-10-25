package core

import "context"

// PortScanner defines the interface for all scanner implementations.
// This interface allows polymorphic handling of TCP and UDP scanners,
// eliminating the need for type switches in client code.
type PortScanner interface {
	// Results returns a read-only channel for receiving scan events.
	Results() <-chan Event

	// ScanRange scans a single host with the specified ports.
	ScanRange(ctx context.Context, host string, ports []uint16)

	// ScanTargets scans multiple targets with their associated ports.
	ScanTargets(ctx context.Context, targets []ScanTarget)
}

// Ensure Scanner implements PortScanner interface
var _ PortScanner = (*Scanner)(nil)
var _ PortScanner = (*UDPScanner)(nil)
