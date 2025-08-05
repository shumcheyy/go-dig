package errors

import (
	"fmt"
	"strings"
	"testing"
)

func TestDigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		digError *DigError
		expected string
	}{
		{
			name: "error without cause",
			digError: &DigError{
				Type:    ErrorTypeInput,
				Message: "invalid domain",
			},
			expected: "Input error: invalid domain",
		},
		{
			name: "error with cause",
			digError: &DigError{
				Type:    ErrorTypeNetwork,
				Message: "connection failed",
				Cause:   fmt.Errorf("timeout"),
			},
			expected: "Network error: connection failed (caused by: timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.digError.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDigError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	digErr := &DigError{
		Type:    ErrorTypeInput,
		Message: "test error",
		Cause:   cause,
	}

	if digErr.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", digErr.Unwrap(), cause)
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		isInput  bool
		isNet    bool
		isDNS    bool
		isSystem bool
	}{
		{
			name:     "input error",
			err:      NewInputError("test", nil),
			isInput:  true,
			isNet:    false,
			isDNS:    false,
			isSystem: false,
		},
		{
			name:     "network error",
			err:      NewNetworkError("test", nil, "8.8.8.8"),
			isInput:  false,
			isNet:    true,
			isDNS:    false,
			isSystem: false,
		},
		{
			name:     "DNS error",
			err:      NewDNSError("test", nil, "example.com", "8.8.8.8"),
			isInput:  false,
			isNet:    false,
			isDNS:    true,
			isSystem: false,
		},
		{
			name:     "system error",
			err:      NewSystemError("test", nil),
			isInput:  false,
			isNet:    false,
			isDNS:    false,
			isSystem: true,
		},
		{
			name:     "non-DigError",
			err:      fmt.Errorf("regular error"),
			isInput:  false,
			isNet:    false,
			isDNS:    false,
			isSystem: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsInputError(tt.err) != tt.isInput {
				t.Errorf("IsInputError() = %v, want %v", IsInputError(tt.err), tt.isInput)
			}
			if IsNetworkError(tt.err) != tt.isNet {
				t.Errorf("IsNetworkError() = %v, want %v", IsNetworkError(tt.err), tt.isNet)
			}
			if IsDNSError(tt.err) != tt.isDNS {
				t.Errorf("IsDNSError() = %v, want %v", IsDNSError(tt.err), tt.isDNS)
			}
			if IsSystemError(tt.err) != tt.isSystem {
				t.Errorf("IsSystemError() = %v, want %v", IsSystemError(tt.err), tt.isSystem)
			}
		})
	}
}

func TestClassifyNetworkError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		server         string
		expectedType   ErrorType
		expectedSubstr string
	}{
		{
			name:           "timeout error",
			err:            fmt.Errorf("context deadline exceeded"),
			server:         "8.8.8.8:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "timeout",
		},
		{
			name:           "connection refused",
			err:            fmt.Errorf("connection refused"),
			server:         "192.168.1.1:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "refused connection",
		},
		{
			name:           "no such host",
			err:            fmt.Errorf("no such host"),
			server:         "invalid.dns.server",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "could not be resolved",
		},
		{
			name:           "network unreachable",
			err:            fmt.Errorf("network unreachable"),
			server:         "10.0.0.1:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "Network unreachable",
		},
		{
			name:           "permission denied",
			err:            fmt.Errorf("permission denied"),
			server:         "8.8.8.8:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "Permission denied",
		},
		{
			name:           "generic network error",
			err:            fmt.Errorf("some other network error"),
			server:         "1.1.1.1:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "Network error occurred",
		},
		{
			name:           "nil error",
			err:            nil,
			server:         "8.8.8.8:53",
			expectedType:   ErrorTypeNetwork,
			expectedSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyNetworkError(tt.err, tt.server)

			if tt.err == nil {
				if result != nil {
					t.Errorf("ClassifyNetworkError() = %v, want nil for nil input", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("ClassifyNetworkError() = nil, want non-nil error")
			}

			if result.Type != tt.expectedType {
				t.Errorf("ClassifyNetworkError().Type = %v, want %v", result.Type, tt.expectedType)
			}

			if result.Server != tt.server {
				t.Errorf("ClassifyNetworkError().Server = %q, want %q", result.Server, tt.server)
			}

			if tt.expectedSubstr != "" && !strings.Contains(result.Message, tt.expectedSubstr) {
				t.Errorf("ClassifyNetworkError().Message = %q, want to contain %q", result.Message, tt.expectedSubstr)
			}

			if result.Cause != tt.err {
				t.Errorf("ClassifyNetworkError().Cause = %v, want %v", result.Cause, tt.err)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid domain",
			domain:    "example.com",
			wantError: false,
		},
		{
			name:      "valid subdomain",
			domain:    "www.example.com",
			wantError: false,
		},
		{
			name:      "valid complex domain",
			domain:    "test-site.example-domain.co.uk",
			wantError: false,
		},
		{
			name:      "empty domain",
			domain:    "",
			wantError: true,
			errorMsg:  "cannot be empty",
		},
		{
			name:      "domain too long",
			domain:    strings.Repeat("a", 254),
			wantError: true,
			errorMsg:  "too long",
		},
		{
			name:      "consecutive dots",
			domain:    "example..com",
			wantError: true,
			errorMsg:  "consecutive dots",
		},
		{
			name:      "starts with dot",
			domain:    ".example.com",
			wantError: true,
			errorMsg:  "start or end with a dot",
		},
		{
			name:      "ends with dot",
			domain:    "example.com.",
			wantError: true,
			errorMsg:  "start or end with a dot",
		},
		{
			name:      "starts with hyphen",
			domain:    "-example.com",
			wantError: true,
			errorMsg:  "start or end with a hyphen",
		},
		{
			name:      "ends with hyphen",
			domain:    "example.com-",
			wantError: true,
			errorMsg:  "start or end with a hyphen",
		},
		{
			name:      "invalid character",
			domain:    "example@.com",
			wantError: true,
			errorMsg:  "invalid character",
		},
		{
			name:      "label too long",
			domain:    strings.Repeat("a", 64) + ".com",
			wantError: true,
			errorMsg:  "too long",
		},
		{
			name:      "label starts with hyphen",
			domain:    "invalid-.com",
			wantError: true,
			errorMsg:  "cannot start or end with hyphen",
		},
		{
			name:      "label ends with hyphen",
			domain:    "invalid-.com",
			wantError: true,
			errorMsg:  "cannot start or end with hyphen",
		},
		{
			name:      "numeric TLD",
			domain:    "example.123",
			wantError: true,
			errorMsg:  "cannot be all numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateDomain() = nil, want error")
					return
				}

				if !IsInputError(err) {
					t.Errorf("ValidateDomain() error type = %T, want *DigError with ErrorTypeInput", err)
				}

				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateDomain() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDomain() = %v, want nil", err)
				}
			}
		})
	}
}

func TestValidateDNSServer(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid IPv4",
			server:    "8.8.8.8",
			wantError: false,
		},
		{
			name:      "valid IPv6",
			server:    "2001:4860:4860::8888",
			wantError: false,
		},
		{
			name:      "localhost IPv4",
			server:    "127.0.0.1",
			wantError: false,
		},
		{
			name:      "localhost IPv6",
			server:    "::1",
			wantError: false,
		},
		{
			name:      "empty server",
			server:    "",
			wantError: true,
			errorMsg:  "cannot be empty",
		},
		{
			name:      "invalid IP format",
			server:    "not.an.ip",
			wantError: true,
			errorMsg:  "not a valid IP address",
		},
		{
			name:      "invalid IPv4",
			server:    "256.256.256.256",
			wantError: true,
			errorMsg:  "not a valid IP address",
		},
		{
			name:      "other loopback address",
			server:    "127.0.0.2",
			wantError: true,
			errorMsg:  "not recommended for DNS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNSServer(tt.server)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateDNSServer() = nil, want error")
					return
				}

				if !IsInputError(err) {
					t.Errorf("ValidateDNSServer() error type = %T, want *DigError with ErrorTypeInput", err)
				}

				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateDNSServer() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDNSServer() = %v, want nil", err)
				}
			}
		})
	}
}

func TestValidateDNSPort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"valid port 53", "53", false},
		{"valid port 8053", "8053", false},
		{"valid port 1", "1", false},
		{"valid port 65535", "65535", false},
		{"empty port", "", true},
		{"port 0", "0", true},
		{"port too high", "65536", true},
		{"port way too high", "99999", true},
		{"non-numeric port", "abc", true},
		{"mixed alphanumeric", "53a", true},
		{"negative port", "-53", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNSPort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDNSPort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}
		})
	}
}

func TestValidateRecordType(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		wantError  bool
		errorMsg   string
	}{
		{
			name:       "valid A record",
			recordType: "A",
			wantError:  false,
		},
		{
			name:       "valid AAAA record",
			recordType: "AAAA",
			wantError:  false,
		},
		{
			name:       "valid MX record",
			recordType: "MX",
			wantError:  false,
		},
		{
			name:       "valid CNAME record",
			recordType: "CNAME",
			wantError:  false,
		},
		{
			name:       "valid TXT record",
			recordType: "TXT",
			wantError:  false,
		},
		{
			name:       "lowercase valid record",
			recordType: "a",
			wantError:  false,
		},
		{
			name:       "mixed case valid record",
			recordType: "Mx",
			wantError:  false,
		},
		{
			name:       "empty record type",
			recordType: "",
			wantError:  true,
			errorMsg:   "cannot be empty",
		},
		{
			name:       "invalid record type",
			recordType: "INVALID",
			wantError:  true,
			errorMsg:   "unsupported record type",
		},
		{
			name:       "numeric record type",
			recordType: "123",
			wantError:  true,
			errorMsg:   "unsupported record type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecordType(tt.recordType)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateRecordType() = nil, want error")
					return
				}

				if !IsInputError(err) {
					t.Errorf("ValidateRecordType() error type = %T, want *DigError with ErrorTypeInput", err)
				}

				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateRecordType() error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRecordType() = %v, want nil", err)
				}
			}
		})
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeInput, "Input"},
		{ErrorTypeNetwork, "Network"},
		{ErrorTypeDNS, "DNS"},
		{ErrorTypeSystem, "System"},
		{ErrorType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
