// Package errors provides user-friendly error types with detailed messages and recovery suggestions.
//
// This package defines custom error types that provide clear, actionable error messages
// to end users. Unlike standard Go errors, these errors include:
//   - Human-readable problem descriptions
//   - Suggested solutions or recovery actions
//   - Context about what the user was trying to do
//
// Example usage:
//
//	if port < 1 || port > 65535 {
//	    return errors.NewUserError(
//	        "Invalid port number",
//	        fmt.Sprintf("Port %d is outside valid range (1-65535)", port),
//	        "Please specify a port between 1 and 65535",
//	    )
//	}
//
// Error Types:
//
// UserError: End-user facing errors with recovery guidance
//   - Problem: What went wrong
//   - Details: Specific information about the error
//   - Solution: How to fix it
//
// Example error messages:
//
//	Error: Invalid CIDR notation
//	Details: "192.168.1.0/33" has invalid subnet mask
//	Solution: CIDR masks must be between /0 and /32 for IPv4
//
//	Error: Target unreachable
//	Details: Host "example.internal" could not be resolved
//	Solution: Check hostname spelling or network connectivity
//
// Pre-defined Errors:
//
// The package includes common error constructors:
//   - ErrInvalidPort: Port outside valid range
//   - ErrInvalidCIDR: Malformed CIDR notation
//   - ErrInvalidHost: Invalid hostname or IP address
//   - ErrLocalhostScanningDisabled: Attempted localhost scan without permission
//   - ErrPrivateIPScanningDisabled: Attempted private IP scan without permission
//
// Integration:
//
// These errors implement the standard error interface, so they work seamlessly
// with Go's error handling:
//
//	if err != nil {
//	    if userErr, ok := err.(*errors.UserError); ok {
//	        fmt.Fprintf(os.Stderr, "Error: %s\n", userErr.Problem)
//	        fmt.Fprintf(os.Stderr, "Details: %s\n", userErr.Details)
//	        fmt.Fprintf(os.Stderr, "Solution: %s\n", userErr.Solution)
//	    }
//	    return err
//	}
package errors
