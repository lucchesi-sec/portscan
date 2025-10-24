package targets

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	minPort         = 1
	maxPort         = 65535
	minRate         = 1
	maxRate         = 15000 // Maximum safe rate in pps
	minTimeout      = 1
	maxTimeout      = 60000
	maxPortsInRange = 65535
)

// ValidateHost validates a host string (IP address, hostname, or CIDR).
// Returns nil if valid, error with details otherwise.
func ValidateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if len(host) > 253 {
		return fmt.Errorf("host exceeds maximum length of 253 characters")
	}

	// Check if it's a valid IP address
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// Check if it's a valid CIDR notation
	if strings.Contains(host, "/") {
		return ValidateCIDR(host)
	}

	// Check if it looks like an IP address but is invalid
	// (contains only dots and digits but failed ParseIP)
	if looksLikeIP(host) {
		return fmt.Errorf("invalid IP address format")
	}

	// Validate as hostname
	return validateHostname(host)
}

// looksLikeIP returns true if the string looks like it's trying to be an IP address
// but failed validation (e.g., "256.1.1.1", "192.168.1", "192.168.1.a")
func looksLikeIP(s string) bool {
	// If it contains dots and starts with a digit, it's likely trying to be an IP
	if !strings.Contains(s, ".") {
		return false
	}

	if len(s) == 0 || s[0] < '0' || s[0] > '9' {
		return false
	}

	// Count dots, digits, and letters
	dotCount := 0
	digitCount := 0
	letterCount := 0
	for _, ch := range s {
		if ch == '.' {
			dotCount++
		} else if ch >= '0' && ch <= '9' {
			digitCount++
		} else if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			letterCount++
		}
	}

	// Only flag as likely invalid IP if it has exactly 3 dots (IPv4 pattern),
	// is >80% digits, and has no letters
	// OR if it has 2-3 dots and is ALL digits (invalid IP missing octets or letters)
	// OR if it has 3 dots and is >60% digits (likely malformed IP like "192.168.1.a")
	if dotCount == 3 && digitCount > len(s)*4/5 && letterCount == 0 {
		return true
	}
	if dotCount >= 2 && dotCount <= 3 && letterCount == 0 && digitCount+dotCount == len(s) {
		return true
	}
	if dotCount == 3 && digitCount > len(s)*3/5 {
		return true
	}
	return false
}

// ValidatePortRange validates a port range specification string.
// Accepts formats: "80", "80,443", "1-1024", "80,443,8000-9000"
func ValidatePortRange(portSpec string) error {
	if portSpec == "" {
		return fmt.Errorf("port specification cannot be empty")
	}

	if len(portSpec) > 1000 {
		return fmt.Errorf("port specification too long (max 1000 characters)")
	}

	tokens := strings.Split(portSpec, ",")
	if len(tokens) == 0 {
		return fmt.Errorf("port specification contains no valid ports")
	}

	validTokenCount := 0
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if err := validatePortToken(token); err != nil {
			return err
		}
		validTokenCount++
	}

	if validTokenCount == 0 {
		return fmt.Errorf("port specification contains no valid ports")
	}

	return nil
}

// ValidateRateLimit validates that the rate limit is within acceptable bounds.
func ValidateRateLimit(rate int) error {
	if rate < minRate {
		return fmt.Errorf("rate limit %d is too low (minimum: %d pps)", rate, minRate)
	}

	if rate > maxRate {
		return fmt.Errorf("rate limit %d is too high (maximum: %d pps)", rate, maxRate)
	}

	return nil
}

// ValidateTimeout validates that the timeout value is within acceptable bounds.
func ValidateTimeout(timeout int) error {
	if timeout < minTimeout {
		return fmt.Errorf("timeout %dms is too low (minimum: %dms)", timeout, minTimeout)
	}

	if timeout > maxTimeout {
		return fmt.Errorf("timeout %dms is too high (maximum: %dms)", timeout, maxTimeout)
	}

	return nil
}

// ValidateCIDR validates CIDR notation and ensures it's not too large.
func ValidateCIDR(cidr string) error {
	if cidr == "" {
		return fmt.Errorf("CIDR notation cannot be empty")
	}

	if !strings.Contains(cidr, "/") {
		return fmt.Errorf("invalid CIDR notation: missing '/' separator")
	}

	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation: %w", err)
	}

	// Verify IP version consistency
	if ip.To4() != nil && network.IP.To4() == nil {
		return fmt.Errorf("inconsistent IP version in CIDR")
	}

	// Check CIDR size to prevent memory exhaustion
	ones, bits := network.Mask.Size()
	if ones == 0 && bits == 0 {
		return fmt.Errorf("invalid CIDR mask")
	}

	// For IPv4, warn about very large ranges
	if bits == 32 {
		hostBits := bits - ones
		if hostBits > 20 { // More than ~1 million hosts
			return fmt.Errorf("CIDR /%d is too large (would generate %d hosts)", ones, 1<<uint(hostBits))
		}
	}

	// For IPv6, be even more restrictive
	if bits == 128 {
		hostBits := bits - ones
		if hostBits > 16 { // More than 65536 hosts
			return fmt.Errorf("IPv6 CIDR /%d is too large (would generate %d hosts)", ones, 1<<uint(hostBits))
		}
	}

	return nil
}

// ValidateWorkers validates the worker count is reasonable.
func ValidateWorkers(workers int) error {
	if workers < 0 {
		return fmt.Errorf("workers cannot be negative")
	}

	// 0 means auto-detect, which is valid
	if workers == 0 {
		return nil
	}

	if workers > 1000 {
		return fmt.Errorf("worker count %d is too high (maximum: 1000)", workers)
	}

	return nil
}

// ValidateUDPWorkerRatio validates the UDP worker ratio is between 0.0 and 1.0.
// -1.0 is allowed as a special value meaning "use default behavior".
func ValidateUDPWorkerRatio(ratio float64) error {
	if ratio == -1.0 {
		return nil // -1.0 means use default behavior
	}
	if ratio < 0.0 || ratio > 1.0 {
		return fmt.Errorf("UDP worker ratio %.2f must be between 0.0 and 1.0 (or -1.0 for default)", ratio)
	}
	return nil
}

// validatePortToken validates a single port or port range token.
func validatePortToken(token string) error {
	if strings.Contains(token, "-") {
		return validateRangeToken(token)
	}
	return validateSinglePort(token)
}

// validateSinglePort validates a single port number.
func validateSinglePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port number '%s': must be numeric", portStr)
	}

	if port < minPort || port > maxPort {
		return fmt.Errorf("port %d out of range (valid: %d-%d)", port, minPort, maxPort)
	}

	return nil
}

// validateRangeToken validates a port range like "80-443".
func validateRangeToken(token string) error {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port range '%s': must be in format 'start-end'", token)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return fmt.Errorf("invalid start port in range '%s': must be numeric", token)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return fmt.Errorf("invalid end port in range '%s': must be numeric", token)
	}

	if start < minPort || start > maxPort {
		return fmt.Errorf("start port %d out of range (valid: %d-%d)", start, minPort, maxPort)
	}

	if end < minPort || end > maxPort {
		return fmt.Errorf("end port %d out of range (valid: %d-%d)", end, minPort, maxPort)
	}

	if start > end {
		return fmt.Errorf("invalid port range '%s': start port %d > end port %d", token, start, end)
	}

	rangeSize := end - start + 1
	if rangeSize > maxPortsInRange {
		return fmt.Errorf("port range '%s' too large (%d ports)", token, rangeSize)
	}

	return nil
}
