package errors

import (
	"fmt"
	"strings"
)

// UserError represents an error with user-friendly message and recovery suggestions
type UserError struct {
	Code       string
	Message    string
	Details    string
	Suggestion string
	WrappedErr error
}

func (e *UserError) Error() string {
	var parts []string

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("Details: %s", e.Details))
	}

	if e.Suggestion != "" {
		parts = append(parts, fmt.Sprintf("Try: %s", e.Suggestion))
	}

	if e.WrappedErr != nil {
		parts = append(parts, fmt.Sprintf("(Error: %v)", e.WrappedErr))
	}

	return strings.Join(parts, "\n")
}

func (e *UserError) Unwrap() error {
	return e.WrappedErr
}

// Common error constructors

func InvalidPortError(port string, err error) *UserError {
	return &UserError{
		Code:       "INVALID_PORT",
		Message:    fmt.Sprintf("Invalid port specification: '%s'", port),
		Details:    "Ports must be between 1 and 65535",
		Suggestion: "Use formats like '80,443' or '1-1024' or '8000-9000'",
		WrappedErr: err,
	}
}

func NoTargetError() *UserError {
	return &UserError{
		Code:       "NO_TARGET",
		Message:    "No target specified",
		Details:    "A target host or network is required for scanning",
		Suggestion: "Provide a target like 'portscan scan 192.168.1.1' or 'portscan scan example.com'",
	}
}

func InvalidTargetError(target string, err error) *UserError {
	return &UserError{
		Code:       "INVALID_TARGET",
		Message:    fmt.Sprintf("Invalid target: '%s'", target),
		Details:    "Could not resolve or parse the target address",
		Suggestion: "Check the hostname/IP is correct. Examples: '192.168.1.1', 'example.com', '10.0.0.0/24'",
		WrappedErr: err,
	}
}

func ConfigLoadError(path string, err error) *UserError {
	return &UserError{
		Code:       "CONFIG_ERROR",
		Message:    "Failed to load configuration",
		Details:    fmt.Sprintf("Could not read config from: %s", path),
		Suggestion: "Run 'portscan config init' to create a default config, or check file permissions",
		WrappedErr: err,
	}
}

func RateLimitError(requested, max int) *UserError {
	return &UserError{
		Code:       "RATE_LIMIT_HIGH",
		Message:    fmt.Sprintf("Rate limit too high: %d pps", requested),
		Details:    fmt.Sprintf("Maximum safe rate is %d packets/second to avoid port exhaustion", max),
		Suggestion: fmt.Sprintf("Use --rate %d or lower. For faster scans, increase --workers instead", max),
	}
}

func NetworkError(operation string, err error) *UserError {
	return &UserError{
		Code:       "NETWORK_ERROR",
		Message:    fmt.Sprintf("Network operation failed: %s", operation),
		Details:    "Check your network connectivity and firewall settings",
		Suggestion: "Try: 1) Test with 'ping' first, 2) Run with --verbose for details, 3) Check firewall rules",
		WrappedErr: err,
	}
}

func PermissionError(operation string) *UserError {
	return &UserError{
		Code:       "PERMISSION_DENIED",
		Message:    fmt.Sprintf("Permission denied for: %s", operation),
		Details:    "This operation requires elevated privileges",
		Suggestion: "Try running with 'sudo' or check your user permissions",
	}
}

func TimeoutError(timeout int) *UserError {
	return &UserError{
		Code:       "TIMEOUT",
		Message:    "Operation timed out",
		Details:    fmt.Sprintf("No response received within %dms", timeout),
		Suggestion: fmt.Sprintf("Try increasing timeout with --timeout %d or check if target is reachable", timeout+100),
	}
}
