package services

import (
	"testing"
)

func TestGetName(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected string
	}{
		// Common TCP ports
		{"SSH", 22, "ssh"},
		{"HTTP", 80, "http"},
		{"HTTPS", 443, "https"},
		{"MySQL", 3306, "mysql"},
		{"PostgreSQL", 5432, "postgresql"},
		{"Redis", 6379, "redis"},
		{"MongoDB", 27017, "mongodb"},
		{"RDP", 3389, "rdp"},
		{"FTP", 21, "ftp"},
		{"Telnet", 23, "telnet"},
		{"SMTP", 25, "smtp"},
		{"POP3", 110, "pop3"},
		{"IMAP", 143, "imap"},
		{"HTTP Alt", 8080, "http-alt"},
		{"HTTPS Alt", 8443, "https-alt"},

		// Common UDP ports
		{"DNS", 53, "dns"},
		{"DHCP Server", 67, "dhcp"},
		{"DHCP Client", 68, "dhcp"},
		{"TFTP", 69, "tftp"},
		{"NTP", 123, "ntp"},
		{"NetBIOS NS", 137, "netbios-ns"},
		{"NetBIOS DGM", 138, "netbios-dgm"},
		{"NetBIOS SSN", 139, "netbios-ssn"},
		{"SNMP", 161, "snmp"},
		{"SNMP Trap", 162, "snmptrap"},
		{"ISAKMP", 500, "isakmp"},
		{"Syslog", 514, "syslog"},
		{"RIP", 520, "rip"},
		{"OpenVPN", 1194, "openvpn"},
		{"L2TP", 1701, "l2tp"},
		{"RADIUS Auth", 1812, "radius"},
		{"RADIUS Acct", 1813, "radius-acct"},
		{"SSDP", 1900, "ssdp"},
		{"STUN", 3478, "stun"},
		{"IPSec NAT", 4500, "ipsec-nat"},
		{"SIP", 5060, "sip"},
		{"SIP TLS", 5061, "sips"},
		{"mDNS", 5353, "mdns"},
		{"LLMNR", 5355, "llmnr"},
		{"Webmin", 10000, "webmin"},
		{"WireGuard", 51820, "wireguard"},

		// Common ports (both TCP and UDP)
		{"SMB", 445, "smb"},

		// Unknown ports
		{"Unknown port 1", 1, "unknown"},
		{"Unknown port 9999", 9999, "unknown"},
		{"Unknown port 12345", 12345, "unknown"},
		{"Unknown port 65534", 65534, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetName(tt.port)
			if result != tt.expected {
				t.Errorf("GetName(%d) = %s; want %s", tt.port, result, tt.expected)
			}
		})
	}
}

func TestGetNameBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected string
	}{
		{"Port 0", 0, "unknown"},
		{"Port 1", 1, "unknown"},
		{"Port 65535 (max)", 65535, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetName(tt.port)
			if result != tt.expected {
				t.Errorf("GetName(%d) = %s; want %s", tt.port, result, tt.expected)
			}
		})
	}
}

func TestGetNameConsistency(t *testing.T) {
	// Test that calling GetName multiple times returns consistent results
	port := uint16(80)
	first := GetName(port)
	second := GetName(port)

	if first != second {
		t.Errorf("GetName(%d) returned inconsistent results: %s vs %s", port, first, second)
	}

	if first != "http" {
		t.Errorf("GetName(%d) = %s; want http", port, first)
	}
}

func TestGetNameAllKnownServices(t *testing.T) {
	// Verify that all known ports return non-"unknown" values
	knownPorts := []uint16{
		// TCP
		21, 22, 23, 25, 80, 110, 143, 443, 3306, 3389, 5432, 6379, 8080, 8443, 27017,
		// UDP
		53, 67, 68, 69, 123, 137, 138, 139, 161, 162, 445, 500, 514, 520,
		1194, 1701, 1812, 1813, 1900, 3478, 4500, 5060, 5061, 5353, 5355, 10000, 51820,
	}

	for _, port := range knownPorts {
		result := GetName(port)
		if result == "unknown" {
			t.Errorf("GetName(%d) returned 'unknown' but should be mapped to a service", port)
		}
		if result == "" {
			t.Errorf("GetName(%d) returned empty string", port)
		}
	}
}

func TestGetNameReturnType(t *testing.T) {
	// Test that GetName always returns a non-empty string
	testPorts := []uint16{0, 1, 22, 80, 443, 9999, 65535}

	for _, port := range testPorts {
		result := GetName(port)
		if result == "" {
			t.Errorf("GetName(%d) returned empty string", port)
		}
	}
}

// Benchmark tests
func BenchmarkGetNameKnownPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetName(80)
	}
}

func BenchmarkGetNameUnknownPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetName(9999)
	}
}

func BenchmarkGetNameMultiplePorts(b *testing.B) {
	ports := []uint16{22, 80, 443, 3306, 5432, 6379, 9999}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, port := range ports {
			GetName(port)
		}
	}
}
