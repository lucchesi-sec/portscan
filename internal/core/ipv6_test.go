package core

import (
	"net"
	"strconv"
	"testing"
)

// TestIPv6AddressConstruction tests that our address construction works for both IPv4 and IPv6
func TestIPv6AddressConstruction(t *testing.T) {
	tests := []struct {
		host     string
		port     uint16
		expected string
	}{
		{"127.0.0.1", 80, "127.0.0.1:80"},
		{"::1", 80, "[::1]:80"},
		{"192.168.1.1", 443, "192.168.1.1:443"},
		{"2001:db8::1", 443, "[2001:db8::1]:443"},
	}

	for _, test := range tests {
		actual := net.JoinHostPort(test.host, strconv.Itoa(int(test.port)))
		if actual != test.expected {
			t.Errorf("For host %s and port %d, expected %s but got %s", test.host, test.port, test.expected, actual)
		}
	}
}
