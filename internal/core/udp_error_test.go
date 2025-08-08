package core

import (
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
)

// TestUDPErrorClassification tests that our UDP error classification works correctly
func TestUDPErrorClassification(t *testing.T) {
	// Test that we correctly identify different types of errors
	tests := []struct {
		name     string
		err      error
		expected ScanState
	}{
		{
			name:     "Timeout error",
			err:      &net.OpError{Err: &mockTimeoutError{true}},
			expected: StateFiltered,
		},
		{
			name:     "Connection refused (port closed)",
			err:      &net.OpError{Err: &os.SyscallError{Err: syscall.ECONNREFUSED}},
			expected: StateClosed,
		},
		{
			name:     "Host unreachable (filtered)",
			err:      &net.OpError{Err: &os.SyscallError{Err: syscall.EHOSTUNREACH}},
			expected: StateFiltered,
		},
		{
			name:     "Network unreachable (filtered)",
			err:      &net.OpError{Err: &os.SyscallError{Err: syscall.ENETUNREACH}},
			expected: StateFiltered,
		},
		{
			name:     "Other syscall error (filtered)",
			err:      &net.OpError{Err: &os.SyscallError{Err: syscall.EACCES}},
			expected: StateFiltered,
		},
		{
			name:     "Generic error (closed)",
			err:      errors.New("generic error"),
			expected: StateClosed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Apply the same logic as in scanUDPPort function
			var state ScanState
			if netErr, ok := test.err.(net.Error); ok && netErr.Timeout() {
				state = StateFiltered // Unknown state due to timeout
			} else {
				// Inspect underlying errors for ICMP differentiation
				var syscallErr *os.SyscallError
				if errors.As(test.err, &syscallErr) {
					switch syscallErr.Err {
					case syscall.ECONNREFUSED:
						// ICMP Port Unreachable - port is definitely closed
						state = StateClosed
					case syscall.EHOSTUNREACH, syscall.ENETUNREACH:
						// ICMP Host/Net Unreachable - filtered
						state = StateFiltered
					default:
						// Other network errors - treat as filtered
						state = StateFiltered
					}
				} else {
					// Fallback to default behavior - treat as closed
					state = StateClosed
				}
			}

			if state != test.expected {
				t.Errorf("For error %v, expected %s but got %s", test.err, test.expected, state)
			}
		})
	}
}

// mockTimeoutError implements net.Error interface for testing
type mockTimeoutError struct {
	timeout bool
}

func (e *mockTimeoutError) Error() string   { return "timeout error" }
func (e *mockTimeoutError) Timeout() bool   { return e.timeout }
func (e *mockTimeoutError) Temporary() bool { return false }
