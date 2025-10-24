package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestUserError_Error tests the Error() method formatting
func TestUserError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *UserError
		contains []string
	}{
		{
			name: "full error with all fields",
			err: &UserError{
				Code:       "TEST_CODE",
				Message:    "test message",
				Details:    "test details",
				Suggestion: "test suggestion",
				WrappedErr: errors.New("wrapped error"),
			},
			contains: []string{"test message", "Details: test details", "Try: test suggestion", "(Error: wrapped error)"},
		},
		{
			name: "error with only message",
			err: &UserError{
				Message: "simple message",
			},
			contains: []string{"simple message"},
		},
		{
			name: "error with message and details",
			err: &UserError{
				Message: "main message",
				Details: "extra details",
			},
			contains: []string{"main message", "Details: extra details"},
		},
		{
			name: "error with message and suggestion",
			err: &UserError{
				Message:    "something failed",
				Suggestion: "try this fix",
			},
			contains: []string{"something failed", "Try: try this fix"},
		},
		{
			name: "error with wrapped error only",
			err: &UserError{
				Message:    "operation failed",
				WrappedErr: fmt.Errorf("underlying cause"),
			},
			contains: []string{"operation failed", "(Error: underlying cause)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Error() = %q, should contain %q", result, expected)
				}
			}
		})
	}
}

// TestUserError_Unwrap tests error unwrapping
func TestUserError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("original error")
	userErr := &UserError{
		Message:    "wrapped",
		WrappedErr: wrappedErr,
	}

	unwrapped := userErr.Unwrap()
	if unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

// TestUserError_UnwrapNil tests unwrapping when no error is wrapped
func TestUserError_UnwrapNil(t *testing.T) {
	userErr := &UserError{
		Message: "no wrapped error",
	}

	unwrapped := userErr.Unwrap()
	if unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

// TestInvalidPortError tests port error creation
func TestInvalidPortError(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		wrappedErr error
	}{
		{
			name:      "invalid port with wrapped error",
			port:      "99999",
			wrappedErr: errors.New("port out of range"),
		},
		{
			name:      "invalid port format",
			port:      "abc",
			wrappedErr: errors.New("not a number"),
		},
		{
			name:      "invalid port without wrapped error",
			port:      "70000",
			wrappedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidPortError(tt.port, tt.wrappedErr)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "INVALID_PORT" {
				t.Errorf("Code = %s, want INVALID_PORT", err.Code)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.port) {
				t.Errorf("Error message should contain port %q", tt.port)
			}

			if !strings.Contains(errMsg, "1 and 65535") {
				t.Error("Error should mention valid port range")
			}

			if tt.wrappedErr != nil && err.WrappedErr != tt.wrappedErr {
				t.Errorf("WrappedErr = %v, want %v", err.WrappedErr, tt.wrappedErr)
			}
		})
	}
}

// TestNoTargetError tests no target error creation
func TestNoTargetError(t *testing.T) {
	err := NoTargetError()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "NO_TARGET" {
		t.Errorf("Code = %s, want NO_TARGET", err.Code)
	}

	errMsg := err.Error()
	expectedPhrases := []string{"No target", "required", "portscan scan"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Error message should contain %q", phrase)
		}
	}
}

// TestInvalidTargetError tests invalid target error creation
func TestInvalidTargetError(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wrappedErr error
	}{
		{
			name:       "invalid hostname",
			target:     "invalid..hostname",
			wrappedErr: errors.New("invalid format"),
		},
		{
			name:       "unresolvable domain",
			target:     "nonexistent.example.local",
			wrappedErr: errors.New("no such host"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidTargetError(tt.target, tt.wrappedErr)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "INVALID_TARGET" {
				t.Errorf("Code = %s, want INVALID_TARGET", err.Code)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.target) {
				t.Errorf("Error message should contain target %q", tt.target)
			}

			if err.WrappedErr != tt.wrappedErr {
				t.Errorf("WrappedErr = %v, want %v", err.WrappedErr, tt.wrappedErr)
			}
		})
	}
}

// TestInvalidTargetListError tests target list error creation
func TestInvalidTargetListError(t *testing.T) {
	wrappedErr := errors.New("multiple resolution failures")
	err := InvalidTargetListError(wrappedErr)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Code != "INVALID_TARGET_LIST" {
		t.Errorf("Code = %s, want INVALID_TARGET_LIST", err.Code)
	}

	errMsg := err.Error()
	expectedPhrases := []string{"Unable to resolve", "targets", "CIDR"}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Error message should contain %q", phrase)
		}
	}

	if err.WrappedErr != wrappedErr {
		t.Errorf("WrappedErr = %v, want %v", err.WrappedErr, wrappedErr)
	}
}

// TestConfigLoadError tests config load error creation
func TestConfigLoadError(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wrappedErr error
	}{
		{
			name:       "file not found",
			path:       "/path/to/config.yaml",
			wrappedErr: errors.New("file not found"),
		},
		{
			name:       "permission denied",
			path:       "/etc/portscan/config.yaml",
			wrappedErr: errors.New("permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConfigLoadError(tt.path, tt.wrappedErr)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "CONFIG_ERROR" {
				t.Errorf("Code = %s, want CONFIG_ERROR", err.Code)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.path) {
				t.Errorf("Error message should contain path %q", tt.path)
			}

			if !strings.Contains(errMsg, "config init") {
				t.Error("Error should mention config init command")
			}

			if err.WrappedErr != tt.wrappedErr {
				t.Errorf("WrappedErr = %v, want %v", err.WrappedErr, tt.wrappedErr)
			}
		})
	}
}

// TestRateLimitError tests rate limit error creation
func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		max       int
	}{
		{
			name:      "rate too high",
			requested: 100000,
			max:       50000,
		},
		{
			name:      "slightly over limit",
			requested: 15001,
			max:       15000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RateLimitError(tt.requested, tt.max)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "RATE_LIMIT_HIGH" {
				t.Errorf("Code = %s, want RATE_LIMIT_HIGH", err.Code)
			}

			errMsg := err.Error()

			// Check that both requested and max values are mentioned
			requestedStr := fmt.Sprintf("%d", tt.requested)
			maxStr := fmt.Sprintf("%d", tt.max)

			if !strings.Contains(errMsg, requestedStr) {
				t.Errorf("Error message should contain requested rate %s", requestedStr)
			}

			if !strings.Contains(errMsg, maxStr) {
				t.Errorf("Error message should contain max rate %s", maxStr)
			}

			if !strings.Contains(errMsg, "workers") {
				t.Error("Error should suggest increasing workers")
			}
		})
	}
}

// TestNetworkError tests network error creation
func TestNetworkError(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		wrappedErr error
	}{
		{
			name:       "connection refused",
			operation:  "connect",
			wrappedErr: errors.New("connection refused"),
		},
		{
			name:       "timeout",
			operation:  "dial",
			wrappedErr: errors.New("i/o timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NetworkError(tt.operation, tt.wrappedErr)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "NETWORK_ERROR" {
				t.Errorf("Code = %s, want NETWORK_ERROR", err.Code)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.operation) {
				t.Errorf("Error message should contain operation %q", tt.operation)
			}

			expectedPhrases := []string{"network", "firewall", "ping"}
			for _, phrase := range expectedPhrases {
				if !strings.Contains(errMsg, phrase) {
					t.Errorf("Error message should contain %q", phrase)
				}
			}

			if err.WrappedErr != tt.wrappedErr {
				t.Errorf("WrappedErr = %v, want %v", err.WrappedErr, tt.wrappedErr)
			}
		})
	}
}

// TestPermissionError tests permission error creation
func TestPermissionError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{
			name:      "raw socket",
			operation: "raw socket creation",
		},
		{
			name:      "bind privileged port",
			operation: "bind to port 80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PermissionError(tt.operation)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "PERMISSION_DENIED" {
				t.Errorf("Code = %s, want PERMISSION_DENIED", err.Code)
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.operation) {
				t.Errorf("Error message should contain operation %q", tt.operation)
			}

			if !strings.Contains(errMsg, "sudo") {
				t.Error("Error should suggest using sudo")
			}
		})
	}
}

// TestTimeoutError tests timeout error creation
func TestTimeoutError(t *testing.T) {
	tests := []struct {
		name    string
		timeout int
	}{
		{
			name:    "200ms timeout",
			timeout: 200,
		},
		{
			name:    "1000ms timeout",
			timeout: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TimeoutError(tt.timeout)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Code != "TIMEOUT" {
				t.Errorf("Code = %s, want TIMEOUT", err.Code)
			}

			errMsg := err.Error()
			timeoutStr := fmt.Sprintf("%dms", tt.timeout)
			if !strings.Contains(errMsg, timeoutStr) {
				t.Errorf("Error message should contain timeout %s", timeoutStr)
			}

			// Check that it suggests increasing timeout
			suggestedTimeout := fmt.Sprintf("%d", tt.timeout+100)
			if !strings.Contains(errMsg, suggestedTimeout) {
				t.Errorf("Error should suggest timeout of %s", suggestedTimeout)
			}
		})
	}
}

// TestUserError_ErrorsAs tests that UserError works with errors.As
func TestUserError_ErrorsAs(t *testing.T) {
	original := &UserError{
		Code:    "TEST",
		Message: "test error",
	}

	var target *UserError
	if !errors.As(original, &target) {
		t.Error("errors.As should work with UserError")
	}

	if target.Code != "TEST" {
		t.Errorf("Code = %s, want TEST", target.Code)
	}
}

// TestUserError_ErrorsIs tests that wrapped errors work with errors.Is
func TestUserError_ErrorsIs(t *testing.T) {
	wrappedErr := errors.New("specific error")
	userErr := &UserError{
		Message:    "wrapper",
		WrappedErr: wrappedErr,
	}

	if !errors.Is(userErr, wrappedErr) {
		t.Error("errors.Is should find wrapped error")
	}
}

// TestErrorConstructors_ReturnNonNil ensures all constructors return valid errors
func TestErrorConstructors_ReturnNonNil(t *testing.T) {
	constructors := []struct {
		name string
		err  *UserError
	}{
		{"InvalidPortError", InvalidPortError("80", nil)},
		{"NoTargetError", NoTargetError()},
		{"InvalidTargetError", InvalidTargetError("test", nil)},
		{"InvalidTargetListError", InvalidTargetListError(errors.New("test"))},
		{"ConfigLoadError", ConfigLoadError("/path", errors.New("test"))},
		{"RateLimitError", RateLimitError(100, 50)},
		{"NetworkError", NetworkError("test", errors.New("test"))},
		{"PermissionError", PermissionError("test")},
		{"TimeoutError", TimeoutError(100)},
	}

	for _, tc := range constructors {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Errorf("%s returned nil", tc.name)
			}

			if tc.err.Code == "" {
				t.Errorf("%s has empty Code", tc.name)
			}

			if tc.err.Message == "" {
				t.Errorf("%s has empty Message", tc.name)
			}

			// Ensure Error() method works
			errStr := tc.err.Error()
			if errStr == "" {
				t.Errorf("%s.Error() returned empty string", tc.name)
			}
		})
	}
}
