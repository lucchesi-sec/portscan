package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lucchesi-sec/portscan/pkg/services"
)

// UDPScanner handles UDP port scanning operations
type UDPScanner struct {
	*Scanner
	serviceProbes map[uint16][]byte     // Service-specific probes for UDP ports
	customProbes  map[uint16][]byte     // Custom user-defined probes
	probeStats    map[uint16]ProbeStats // Statistics for probe effectiveness
}

// ProbeStats tracks the effectiveness of probes for each port
type ProbeStats struct {
	Sent      int // Number of probes sent
	Responses int // Number of responses received
	Successes int // Number of successful service detections
}

// NewUDPScanner creates a new UDP scanner instance
func NewUDPScanner(cfg *Config) *UDPScanner {
	return &UDPScanner{
		Scanner:       NewScanner(cfg),
		serviceProbes: initUDPProbes(),
		customProbes:  make(map[uint16][]byte),
		probeStats:    make(map[uint16]ProbeStats),
	}
}

// initUDPProbes initializes service-specific UDP probes
func initUDPProbes() map[uint16][]byte {
	return map[uint16][]byte{
		53:    buildDNSProbe(),                // DNS
		123:   buildNTPProbe(),                // NTP
		161:   buildSNMPProbe(),               // SNMP
		500:   buildIKEProbe(),                // IKE/IPSec
		1194:  {0x38, 0x01, 0x00, 0x00, 0x00}, // OpenVPN
		51820: {0x01, 0x00, 0x00, 0x00},       // WireGuard
		67:    buildDHCPProbe(),               // DHCP
		69:    buildTFTPProbe(),               // TFTP
		137:   buildNetBIOSProbe(),            // NetBIOS Name Service
		5353:  buildMDNSProbe(),               // mDNS
	}
}

// scanUDPPort performs UDP port scanning
func (s *UDPScanner) scanUDPPort(ctx context.Context, host string, port uint16) {
	start := time.Now()
	address := net.JoinHostPort(host, strconv.Itoa(int(port)))

	// Create UDP connection with context support
	dialer := &net.Dialer{Timeout: s.config.Timeout}
	conn, err := dialer.DialContext(ctx, "udp", address)
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return
		}

		// Record probe attempt with failure
		s.recordProbeAttempt(port, false)

		// UDP "connection" rarely fails - this indicates network issues
		result := ResultEvent{
			Host:     host,
			Port:     port,
			State:    StateFiltered,
			Protocol: "udp",
			Duration: time.Since(start),
		}
		select {
		case s.results <- result:
			s.completed.Add(1)
		case <-ctx.Done():
		}
		return
	}
	defer func() { _ = conn.Close() }()

	// Set read deadline
	_ = conn.SetReadDeadline(time.Now().Add(s.config.UDPReadTimeout))

	// Send service-specific probe if available, otherwise send empty packet
	probe := s.getProbeForPort(port)
	_, err = conn.Write(probe)
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return
		}

		// Record probe attempt with failure
		s.recordProbeAttempt(port, false)

		result := ResultEvent{
			Host:     host,
			Port:     port,
			State:    StateFiltered,
			Protocol: "udp",
			Duration: time.Since(start),
		}
		select {
		case s.results <- result:
			s.completed.Add(1)
		case <-ctx.Done():
		}
		return
	}

	// Try to read response
	buffer := make([]byte, s.config.UDPBufferSize)
	n, err := conn.Read(buffer)

	// Check if it's a context cancellation before processing results
	if ctx.Err() != nil {
		return
	}

	duration := time.Since(start)
	result := ResultEvent{
		Host:     host,
		Port:     port,
		Protocol: "udp",
		Duration: duration,
	}

	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return
		}

		// Record probe attempt - timeout doesn't count as response
		s.recordProbeAttempt(port, false)

		// Timeout means unknown state (filtered)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.State = StateFiltered // Unknown state due to timeout
		} else {
			// Inspect underlying errors for ICMP differentiation
			var syscallErr *os.SyscallError
			if errors.As(err, &syscallErr) {
				switch syscallErr.Err {
				case syscall.ECONNREFUSED:
					// ICMP Port Unreachable - port is definitely closed
					result.State = StateClosed
				case syscall.EHOSTUNREACH, syscall.ENETUNREACH:
					// ICMP Host/Net Unreachable - filtered
					result.State = StateFiltered
				default:
					// Other network errors - treat as filtered
					result.State = StateFiltered
				}
			} else {
				// Fallback to default behavior - treat as closed
				result.State = StateClosed
			}
		}
	} else {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return
		}

		// Record probe attempt with success
		s.recordProbeAttempt(port, true)

		// Got a response - port is definitely open
		result.State = StateOpen
		if n > 0 && s.config.BannerGrab {
			result.Banner = s.parseUDPResponse(port, buffer[:n])
		}
	}

	select {
	case s.results <- result:
		s.completed.Add(1)
	case <-ctx.Done():
	}
}

// getProbeForPort returns the appropriate probe for a given port
func (s *UDPScanner) getProbeForPort(port uint16) []byte {
	// Check for custom probe first
	if probe, exists := s.customProbes[port]; exists {
		return probe
	}

	// Check for service-specific probe
	if probe, exists := s.serviceProbes[port]; exists {
		return probe
	}

	// Default: send empty UDP packet
	return []byte{}
}

// AddCustomProbe adds a custom probe for a specific port
func (s *UDPScanner) AddCustomProbe(port uint16, probe []byte) {
	s.customProbes[port] = probe
}

// GetProbeStats returns statistics for probe effectiveness
func (s *UDPScanner) GetProbeStats() map[uint16]ProbeStats {
	return s.probeStats
}

// recordProbeAttempt records statistics for a probe attempt
func (s *UDPScanner) recordProbeAttempt(port uint16, success bool) {
	stats := s.probeStats[port]
	stats.Sent++
	if success {
		stats.Responses++
		stats.Successes++
	}
	s.probeStats[port] = stats
}

// parseUDPResponse interprets UDP response based on port/service with confidence scoring
func (s *UDPScanner) parseUDPResponse(port uint16, data []byte) string {
	var service string
	var confidence float64

	switch port {
	case 53: // DNS
		service, confidence = parseDNSResponseWithConfidence(data)
	case 123: // NTP
		service, confidence = parseNTPResponseWithConfidence(data)
	case 161: // SNMP
		service, confidence = parseSNMPResponseWithConfidence(data)
	case 67, 68: // DHCP
		service, confidence = "DHCP", 0.8
	case 137: // NetBIOS
		service, confidence = parseNetBIOSResponseWithConfidence(data)
	case 1194: // OpenVPN
		if len(data) > 0 {
			service, confidence = "OpenVPN", 0.7
		} else {
			service, confidence = "OpenVPN (no response)", 0.3
		}
	case 500, 4500: // IPSec
		service, confidence = "IKE/IPSec", 0.8
	case 51820: // WireGuard
		service, confidence = "WireGuard", 0.7
	case 5353: // mDNS
		service, confidence = "mDNS/Bonjour", 0.8
	default:
		// For unknown services, return sanitized banner
		banner := string(data)
		// Remove non-printable characters and handle Unicode properly
		banner = strings.Map(func(r rune) rune {
			// Keep printable ASCII characters and common Unicode printable characters
			if r >= 32 && r < 127 {
				return r
			}
			// Keep common Unicode printable characters (some ranges)
			if (r >= 160 && r <= 591) || (r >= 880 && r <= 1023) || (r >= 1024 && r <= 1279) {
				return r
			}
			// Remove all other characters
			return -1
		}, banner)

		if len(banner) > 0 {
			if len(banner) > 64 {
				banner = banner[:64] + "..."
			}
			// Add confidence based on banner length and content
			// Formula: 0.5 + (banner_length / 100), capped at 0.9
			// This gives a base confidence of 50% that increases with banner length,
			// but caps at 90% to avoid overconfidence in longer but still uncertain responses
			confidence = math.Min(0.5+float64(len(banner))/100.0, 0.9)
			return fmt.Sprintf("%s (%.1f%%)", banner, confidence*100)
		}

		// If no banner and no specific parser, return tentative service name with "?" prefix
		// This provides useful information to the user even when we can't definitively identify the service
		serviceName := services.GetName(port)
		if serviceName != "unknown" {
			service, confidence = "?"+serviceName, 0.3
		} else {
			service, confidence = "", 0.0
		}
	}

	// Format the result with confidence if we have a service
	if service != "" {
		return fmt.Sprintf("%s (%.1f%%)", service, confidence*100)
	}
	return ""
}

// UDP probe builders

func buildDNSProbe() []byte {
	// DNS query for version.bind TXT (commonly responds)
	return []byte{
		0x00, 0x00, // Transaction ID
		0x01, 0x00, // Flags: standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, // Answer RRs: 0
		0x00, 0x00, // Authority RRs: 0
		0x00, 0x00, // Additional RRs: 0
		// Query: version.bind
		0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
		0x04, 0x62, 0x69, 0x6e, 0x64,
		0x00,       // Root domain
		0x00, 0x10, // Type: TXT
		0x00, 0x03, // Class: CH (Chaos)
	}
}

func buildNTPProbe() []byte {
	// NTP version 3, client mode
	return []byte{
		0x1b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func buildSNMPProbe() []byte {
	// SNMPv1 GetRequest for sysDescr
	return []byte{
		0x30, 0x26, // SEQUENCE
		0x02, 0x01, 0x00, // Version: 1
		0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, // Community: "public"
		0xa0, 0x19, // GetRequest PDU
		0x02, 0x01, 0x00, // Request ID
		0x02, 0x01, 0x00, // Error status
		0x02, 0x01, 0x00, // Error index
		0x30, 0x0e, // Varbind list
		0x30, 0x0c, // Varbind
		0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00, // OID: sysDescr
		0x05, 0x00, // Value: NULL
	}
}

func buildIKEProbe() []byte {
	// IKE version 1 main mode init
	return []byte{
		// IKE Header
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Initiator cookie
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Responder cookie
		0x01,                   // Next payload
		0x10,                   // Version
		0x02,                   // Exchange type: Identity Protection (Main Mode)
		0x00,                   // Flags
		0x00, 0x00, 0x00, 0x00, // Message ID
		0x00, 0x00, 0x00, 0x1c, // Length
	}
}

func buildDHCPProbe() []byte {
	// DHCP Discover message (minimal)
	probe := make([]byte, 240)
	probe[0] = 0x01 // Boot request
	probe[1] = 0x01 // Ethernet
	probe[2] = 0x06 // Hardware address length
	return probe
}

func buildTFTPProbe() []byte {
	// TFTP Read Request for a non-existent file
	return []byte{0x00, 0x01, 0x74, 0x65, 0x73, 0x74, 0x00, 0x6f, 0x63, 0x74, 0x65, 0x74, 0x00}
}

func buildNetBIOSProbe() []byte {
	// NetBIOS Name Service query
	return []byte{
		0x00, 0x00, // Transaction ID
		0x00, 0x10, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x20, 0x43, 0x4b, 0x41, 0x41, 0x41, 0x41, 0x41, // Encoded name
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x00,
		0x00, 0x21, // Type: NB
		0x00, 0x01, // Class: IN
	}
}

func buildMDNSProbe() []byte {
	// mDNS query for _services._dns-sd._udp.local
	return []byte{
		0x00, 0x00, // Transaction ID
		0x00, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x09, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, // _services
		0x07, 0x5f, 0x64, 0x6e, 0x73, 0x2d, 0x73, 0x64, // _dns-sd
		0x04, 0x5f, 0x75, 0x64, 0x70, // _udp
		0x05, 0x6c, 0x6f, 0x63, 0x61, 0x6c, // local
		0x00,       // Root
		0x00, 0x0c, // Type: PTR
		0x00, 0x01, // Class: IN
	}
}

// Response parsers

// parseDNSResponse is commented out for future use
// func parseDNSResponse(data []byte) string {
// 	if len(data) > 12 {
// 		return "DNS"
// 	}
// 	return "DNS (no response)"
// }

func parseDNSResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 12 {
		// Check for valid DNS header
		if len(data) >= 4 && data[2]&0x80 != 0 { // QR bit set (response)
			return "DNS", 0.95
		}
		return "DNS", 0.8
	}
	return "DNS (no response)", 0.3
}

// parseNTPResponse is commented out for future use
// func parseNTPResponse(data []byte) string {
// 	if len(data) >= 48 {
// 		return "NTP"
// 	}
// 	return "NTP (invalid response)"
// }

func parseNTPResponseWithConfidence(data []byte) (string, float64) {
	if len(data) >= 48 {
		// Check for valid NTP fields
		liVnMode := data[0]
		version := (liVnMode >> 3) & 0x07
		mode := liVnMode & 0x07

		// Valid NTP versions are 2-4, mode 4 is server
		if version >= 2 && version <= 4 && mode == 4 {
			return "NTP", 0.95
		}
		return "NTP", 0.8
	}
	return "NTP (invalid response)", 0.2
}

// parseSNMPResponse is commented out for future use
// func parseSNMPResponse(data []byte) string {
// 	if len(data) > 0 && data[0] == 0x30 {
// 		return "SNMP"
// 	}
// 	return "SNMP (no response)"
// }

func parseSNMPResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 0 {
		// Check for ASN.1 SEQUENCE (0x30)
		if data[0] == 0x30 {
			// Check for minimum SNMP structure
			if len(data) >= 3 && data[1] < 0x80 {
				return "SNMP", 0.9
			}
			return "SNMP", 0.7
		}
		// Check for SNMP error responses
		if data[0] == 0xA0 || data[0] == 0xA1 || data[0] == 0xA2 {
			return "SNMP (error)", 0.8
		}
	}
	return "SNMP (no response)", 0.2
}

// parseNetBIOSResponse is commented out for future use
// func parseNetBIOSResponse(data []byte) string {
// 	if len(data) > 12 {
// 		return "NetBIOS Name Service"
// 	}
// 	return "NetBIOS"
// }

func parseNetBIOSResponseWithConfidence(data []byte) (string, float64) {
	if len(data) > 12 {
		// Check for valid NetBIOS header
		if len(data) >= 4 {
			// Check transaction ID and flags
			flags := uint16(data[2])<<8 | uint16(data[3])
			// Response bit should be set
			if flags&0x8000 != 0 {
				return "NetBIOS Name Service", 0.9
			}
		}
		return "NetBIOS Name Service", 0.7
	}
	return "NetBIOS", 0.3
}

// ScanRangeUDP performs UDP scanning on the specified ports
func (s *UDPScanner) ScanRangeUDP(ctx context.Context, host string, ports []uint16) {
	jobs := make(chan uint16, len(ports))

	// Start progress reporter
	progressDone := make(chan struct{})
	go func() {
		s.progressReporter(ctx, len(ports))
		close(progressDone)
	}()

	// Start workers - fewer workers for UDP due to ICMP rate limiting
	var workers int

	// Handle worker ratio:
	// < 0 = use default (half of TCP workers)
	// 0 = explicitly disable UDP workers (use minimal workers)
	// 0.1-1.0 = use this ratio of TCP workers
	if s.config.UDPWorkerRatio < 0 {
		// Default behavior: use half of TCP workers
		workers = s.config.Workers / 2
	} else if s.config.UDPWorkerRatio == 0 {
		// Explicitly set to 0 - use minimal workers to avoid disabling UDP scanning entirely
		workers = 1
	} else {
		// Use the specified ratio
		workers = int(float64(s.config.Workers) * s.config.UDPWorkerRatio)
	}

	// Ensure at least 1 worker to prevent completely disabling UDP scanning
	if workers < 1 {
		workers = 1
	}

	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.udpWorker(ctx, host, jobs)
	}

	// Feed jobs
	go func() {
		for _, port := range ports {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- port:
			}
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	s.wg.Wait()

	// Wait for progress reporter to finish or context to be cancelled
	select {
	case <-progressDone:
	case <-ctx.Done():
		// Context was cancelled, don't block indefinitely
	}

	// Stop rate ticker if used
	if s.rateTicker != nil {
		s.rateTicker.Stop()
	}

	// Now safe to close results
	close(s.results)
}

// ScanRange implements the Scanner interface for UDP scanning
func (s *UDPScanner) ScanRange(ctx context.Context, host string, ports []uint16) {
	s.ScanRangeUDP(ctx, host, ports)
}

func (s *UDPScanner) udpWorker(ctx context.Context, host string, jobs <-chan uint16) {
	defer s.wg.Done()

	// Create a random number generator for jitter
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			return
		case port, ok := <-jobs:
			if !ok {
				return
			}

			// Rate limit if enabled (more conservative for UDP)
			if s.rateTicker != nil {
				select {
				case <-ctx.Done():
					return
				case <-s.rateTicker.C:
					// For UDP, add a small random jitter to avoid ICMP rate limits
					// Jitter between 0-UDPJitterMaxMs ms
					jitter := time.Duration(rng.Intn(s.config.UDPJitterMaxMs)) * time.Millisecond
					time.Sleep(jitter)
				}
			}

			s.scanUDPPort(ctx, host, port)
		}
	}
}
