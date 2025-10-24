package targets

import (
	"strings"
	"testing"
)

// TestValidateHost tests host validation with various inputs
func TestValidateHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		// Valid IPs
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv4 localhost", "127.0.0.1", false},
		{"valid IPv4 boundary", "255.255.255.255", false},
		{"valid IPv6", "2001:db8::1", false},
		{"valid IPv6 localhost", "::1", false},
		{"valid IPv6 full", "2001:0db8:0000:0000:0000:ff00:0042:8329", false},

		// Valid hostnames
		{"valid hostname simple", "example.com", false},
		{"valid hostname subdomain", "www.example.com", false},
		{"valid hostname deep", "api.v2.example.com", false},
		{"valid hostname with hyphen", "my-server.example.com", false},
		{"valid hostname single char", "a.com", false},
		{"valid hostname numbers", "server1.example.com", false},
		{"valid hostname localhost", "localhost", false},
		{"valid hostname digit-heavy", "192168.com", false},
		{"valid hostname mostly digits", "12345.org", false},
		{"valid hostname numeric prefix", "999.example.com", false},

		// Valid CIDR
		{"valid CIDR /24", "192.168.1.0/24", false},
		{"valid CIDR /32", "10.0.0.1/32", false},
		{"valid CIDR /16", "172.16.0.0/16", false},
		{"valid IPv6 CIDR", "2001:db8::/112", false}, // Changed to /112 (65k hosts)

		// Invalid - empty or too long
		{"empty host", "", true},
		{"too long hostname", strings.Repeat("a", 254), true},

		// Invalid IPs
		{"invalid IP octets", "256.1.1.1", true},
		{"invalid IP format", "192.168.1", true},
		{"invalid IP letters", "192.168.1.a", true},

		// Invalid hostnames
		{"hostname starts with hyphen", "-example.com", true},
		{"hostname ends with hyphen", "example.com-", true},
		{"hostname starts with dot", ".example.com", true},
		{"hostname ends with dot", "example.com.", true},
		{"hostname double dot", "example..com", true},
		{"hostname invalid char", "exam ple.com", true},
		{"hostname invalid char @", "exam@ple.com", true},
		{"hostname label too long", strings.Repeat("a", 64) + ".com", true},

		// Invalid CIDR
		{"invalid CIDR no mask", "192.168.1.0/", true},
		{"invalid CIDR bad IP", "256.1.1.1/24", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}

// TestValidatePortRange tests port range validation
func TestValidatePortRange(t *testing.T) {
	tests := []struct {
		name    string
		portSpec string
		wantErr bool
	}{
		// Valid single ports
		{"single port 80", "80", false},
		{"single port 443", "443", false},
		{"single port min", "1", false},
		{"single port max", "65535", false},

		// Valid port lists
		{"multiple ports", "80,443,8080", false},
		{"ports with spaces", "80, 443, 8080", false},
		{"mixed format", "80,443,8000-9000", false},

		// Valid ranges
		{"simple range", "1-1024", false},
		{"web range", "80-443", false},
		{"full range", "1-65535", false},
		{"single port range", "80-80", false},

		// Invalid - empty or malformed
		{"empty spec", "", true},
		{"too long spec", strings.Repeat("1,", 501), true},

		// Invalid ports
		{"port zero", "0", true},
		{"port too high", "65536", true},
		{"port negative", "-1", true},
		{"port non-numeric", "abc", true},
		{"port with letters", "80a", true},

		// Invalid ranges
		{"range backwards", "443-80", true},
		{"range start zero", "0-1024", true},
		{"range end too high", "1-65536", true},
		{"range missing end", "80-", true},
		{"range missing start", "-443", true},
		{"range multiple hyphens", "80--443", true},
		{"range non-numeric start", "abc-443", true},
		{"range non-numeric end", "80-xyz", true},

		// Edge cases
		{"comma only", ",", true},
		{"multiple commas", "80,,443", false}, // Empty tokens are skipped
		{"trailing comma", "80,443,", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortRange(tt.portSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortRange(%q) error = %v, wantErr %v", tt.portSpec, err, tt.wantErr)
			}
		})
	}
}

// TestValidateRateLimit tests rate limit validation
func TestValidateRateLimit(t *testing.T) {
	tests := []struct {
		name    string
		rate    int
		wantErr bool
	}{
		// Valid rates
		{"minimum rate", 1, false},
		{"low rate", 100, false},
		{"default rate", 7500, false},
		{"high rate", 10000, false},
		{"maximum rate", 15000, false},

		// Invalid rates
		{"zero rate", 0, true},
		{"negative rate", -1, true},
		{"negative large", -1000, true},
		{"too high", 15001, true},
		{"way too high", 100000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimit(tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRateLimit(%d) error = %v, wantErr %v", tt.rate, err, tt.wantErr)
			}
		})
	}
}

// TestValidateTimeout tests timeout validation
func TestValidateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		wantErr bool
	}{
		// Valid timeouts
		{"minimum timeout", 1, false},
		{"very short timeout", 10, false},
		{"default timeout", 200, false},
		{"long timeout", 5000, false},
		{"maximum timeout", 60000, false},

		// Invalid timeouts
		{"zero timeout", 0, true},
		{"negative timeout", -1, true},
		{"negative large", -1000, true},
		{"too high", 60001, true},
		{"way too high", 120000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeout(tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeout(%d) error = %v, wantErr %v", tt.timeout, err, tt.wantErr)
			}
		})
	}
}

// TestValidateCIDR tests CIDR validation
func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
	}{
		// Valid CIDRs
		{"valid /32", "192.168.1.1/32", false},
		{"valid /24", "192.168.1.0/24", false},
		{"valid /16", "172.16.0.0/16", false},
		{"valid /12", "10.0.0.0/12", false},
		{"valid IPv6 /112", "2001:db8::/112", false},
		{"valid IPv6 /120", "2001:db8::/120", false},
		{"valid IPv6 /128", "2001:db8::1/128", false},

		// Invalid - empty or malformed
		{"empty CIDR", "", true},
		{"no slash", "192.168.1.0", true},
		{"no mask", "192.168.1.0/", true},
		{"invalid IP", "256.1.1.1/24", true},
		{"invalid mask", "192.168.1.0/33", true},
		{"negative mask", "192.168.1.0/-1", true},

		// Too large ranges
		{"IPv4 /8 too large", "10.0.0.0/8", true},
		{"IPv4 /0 too large", "0.0.0.0/0", true},
		{"IPv6 /32 too large", "2001:db8::/32", true},
		{"IPv6 /64 too large", "2001:db8::/64", true},
		{"IPv6 /0 too large", "::/0", true},

		// Boundary cases
		{"IPv4 /20 boundary", "10.0.0.0/20", false},
		{"IPv6 /116 boundary", "2001:db8::/116", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCIDR(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
			}
		})
	}
}

// TestValidateWorkers tests worker count validation
func TestValidateWorkers(t *testing.T) {
	tests := []struct {
		name    string
		workers int
		wantErr bool
	}{
		// Valid worker counts
		{"auto-detect", 0, false},
		{"minimum workers", 1, false},
		{"low workers", 10, false},
		{"default workers", 100, false},
		{"high workers", 500, false},
		{"maximum workers", 1000, false},

		// Invalid worker counts
		{"negative workers", -1, true},
		{"negative large", -100, true},
		{"too many workers", 1001, true},
		{"way too many", 10000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkers(tt.workers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkers(%d) error = %v, wantErr %v", tt.workers, err, tt.wantErr)
			}
		})
	}
}

// TestValidateUDPWorkerRatio tests UDP worker ratio validation
func TestValidateUDPWorkerRatio(t *testing.T) {
	tests := []struct {
		name    string
		ratio   float64
		wantErr bool
	}{
		// Valid ratios
		{"default ratio", -1.0, false}, // -1.0 means use default
		{"minimum ratio", 0.0, false},
		{"low ratio", 0.1, false},
		{"half ratio", 0.5, false},
		{"high ratio", 0.9, false},
		{"maximum ratio", 1.0, false},

		// Invalid ratios
		{"negative ratio", -0.1, true},
		{"negative large", -2.0, true},
		{"too high", 1.1, true},
		{"way too high", 10.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUDPWorkerRatio(tt.ratio)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUDPWorkerRatio(%f) error = %v, wantErr %v", tt.ratio, err, tt.wantErr)
			}
		})
	}
}

// TestValidateHost_Malicious tests validation against potentially malicious inputs
func TestValidateHost_Malicious(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		// Injection attempts
		{"SQL injection attempt", "'; DROP TABLE hosts; --", true},
		{"command injection", "localhost; rm -rf /", true},
		{"path traversal", "../../../etc/passwd", true},
		{"null byte", "example.com\x00.evil.com", true},
		{"unicode tricks", "еxample.com", true}, // Cyrillic 'e' - reject for security

		// Buffer overflow attempts
		{"very long hostname", strings.Repeat("a", 300), true},
		{"many labels", strings.Repeat("a.", 100) + "com", false}, // Valid if under 253 chars (203 chars)

		// Protocol confusion
		{"http scheme", "http://example.com", true},
		{"https scheme", "https://example.com", true},
		{"ftp scheme", "ftp://example.com", true},

		// Special characters
		{"backslash", "example\\test.com", true},
		{"forward slash", "example/test.com", true},
		{"at symbol", "user@example.com", true},
		{"colon", "example:8080", true},
		{"semicolon", "example;test", true},
		{"quote", "example'test", true},
		{"double quote", "example\"test", true},
		{"backtick", "example`test", true},
		{"pipe", "example|test", true},
		{"ampersand", "example&test", true},
		{"question mark", "example?test", true},
		{"hash", "example#test", true},
		{"percent", "example%test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}

// TestValidatePortRange_EdgeCases tests edge cases for port range validation
func TestValidatePortRange_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		portSpec string
		wantErr bool
	}{
		// Boundary values
		{"port 1", "1", false},
		{"port 65535", "65535", false},
		{"range 1-65535", "1-65535", false},

		// Whitespace handling
		{"leading space", " 80", false},
		{"trailing space", "80 ", false},
		{"spaces around range", " 80 - 443 ", false},
		{"tabs", "\t80\t", false},

		// Malformed but parseable
		{"multiple commas between", "80,,,443", false}, // Empty tokens skipped
		{"spaces and commas", "80 , 443 , 8080", false},

		// Malicious inputs
		{"extremely long list", strings.Repeat("80,", 200), false}, // 600 chars, under 1000 limit
		{"integer overflow attempt", "999999999999999999", true},
		{"negative in range", "-80-443", true},
		{"special chars", "80;443", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortRange(tt.portSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortRange(%q) error = %v, wantErr %v", tt.portSpec, err, tt.wantErr)
			}
		})
	}
}

// TestValidateRateLimit_Boundaries tests boundary conditions
func TestValidateRateLimit_Boundaries(t *testing.T) {
	tests := []struct {
		name    string
		rate    int
		wantErr bool
	}{
		{"just below minimum", 0, true},
		{"at minimum", 1, false},
		{"just above minimum", 2, false},
		{"just below maximum", 14999, false},
		{"at maximum", 15000, false},
		{"just above maximum", 15001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimit(tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRateLimit(%d) error = %v, wantErr %v", tt.rate, err, tt.wantErr)
			}
		})
	}
}

// TestValidateTimeout_Boundaries tests boundary conditions
func TestValidateTimeout_Boundaries(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
		wantErr bool
	}{
		{"just below minimum", 0, true},
		{"at minimum", 1, false},
		{"just above minimum", 2, false},
		{"just below maximum", 59999, false},
		{"at maximum", 60000, false},
		{"just above maximum", 60001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeout(tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeout(%d) error = %v, wantErr %v", tt.timeout, err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkValidateHost(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateHost("example.com")
	}
}

func BenchmarkValidateHostIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateHost("192.168.1.1")
	}
}

func BenchmarkValidatePortRange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidatePortRange("1-1024")
	}
}

func BenchmarkValidateCIDR(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ValidateCIDR("192.168.1.0/24")
	}
}
