package cmd

import (
	"go-dig/pkg/errors"
	"strings"
	"testing"
	"time"
)

func TestCLIParser_Parse_ValidInputs(t *testing.T) {
	parser := NewCLIParser()

	tests := []struct {
		name     string
		args     []string
		expected *Config
	}{
		{
			name: "basic domain only",
			args: []string{"google.com"},
			expected: &Config{
				Domain:     "google.com",
				RecordType: "A",
				Server:     "",
				Timeout:    5 * time.Second,
			},
		},
		{
			name: "domain with record type",
			args: []string{"-t", "AAAA", "google.com"},
			expected: &Config{
				Domain:     "google.com",
				RecordType: "AAAA",
				Server:     "",
				Timeout:    5 * time.Second,
			},
		},
		{
			name: "domain with DNS server",
			args: []string{"-s", "8.8.8.8", "google.com"},
			expected: &Config{
				Domain:     "google.com",
				RecordType: "A",
				Server:     "8.8.8.8",
				Timeout:    5 * time.Second,
			},
		},
		{
			name: "all options",
			args: []string{"-t", "MX", "-s", "1.1.1.1", "example.org"},
			expected: &Config{
				Domain:     "example.org",
				RecordType: "MX",
				Server:     "1.1.1.1",
				Timeout:    5 * time.Second,
			},
		},
		{
			name: "lowercase record type (should be converted to uppercase)",
			args: []string{"-t", "cname", "test.com"},
			expected: &Config{
				Domain:     "test.com",
				RecordType: "CNAME",
				Server:     "",
				Timeout:    5 * time.Second,
			},
		},
		{
			name: "subdomain",
			args: []string{"sub.domain.example.com"},
			expected: &Config{
				Domain:     "sub.domain.example.com",
				RecordType: "A",
				Server:     "",
				Timeout:    5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}

			if config.Domain != tt.expected.Domain {
				t.Errorf("Domain = %v, want %v", config.Domain, tt.expected.Domain)
			}
			if config.RecordType != tt.expected.RecordType {
				t.Errorf("RecordType = %v, want %v", config.RecordType, tt.expected.RecordType)
			}
			if config.Server != tt.expected.Server {
				t.Errorf("Server = %v, want %v", config.Server, tt.expected.Server)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout = %v, want %v", config.Timeout, tt.expected.Timeout)
			}
		})
	}
}

func TestCLIParser_Parse_InvalidInputs(t *testing.T) {
	parser := NewCLIParser()

	tests := []struct {
		name        string
		args        []string
		expectError string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: "domain name is required",
		},
		{
			name:        "too many arguments",
			args:        []string{"google.com", "extra.com"},
			expectError: "too many arguments",
		},
		{
			name:        "invalid record type",
			args:        []string{"-t", "INVALID", "google.com"},
			expectError: "unsupported record type",
		},
		{
			name:        "invalid DNS server IP",
			args:        []string{"-s", "invalid-ip", "google.com"},
			expectError: "not a valid IP address",
		},
		{
			name:        "empty DNS server",
			args:        []string{"-s", "", "google.com"},
			expectError: "cannot be empty",
		},
		{
			name:        "invalid domain - empty",
			args:        []string{""},
			expectError: "cannot be empty",
		},
		{
			name:        "invalid domain - starts with dot",
			args:        []string{".google.com"},
			expectError: "start or end with a dot",
		},
		{
			name:        "invalid domain - ends with dot",
			args:        []string{"google.com."},
			expectError: "start or end with a dot",
		},
		{
			name:        "invalid domain - consecutive dots",
			args:        []string{"google..com"},
			expectError: "consecutive dots",
		},
		{
			name:        "invalid domain - starts with hyphen",
			args:        []string{"-google.com"},
			expectError: "invalid command line arguments",
		},
		{
			name:        "invalid domain - invalid characters",
			args:        []string{"google@com"},
			expectError: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.args)
			if err == nil {
				t.Fatalf("Parse() error = nil, want error containing %q", tt.expectError)
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Parse() error = %v, want error containing %q", err, tt.expectError)
			}

			// Verify that input errors are properly typed
			if !strings.Contains(tt.name, "flag parsing error") && !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	validDomains := []string{
		"google.com",
		"sub.domain.example.org",
		"a.b",
		"test-domain.com",
		"xn--nxasmq6b", // IDN domain
	}

	for _, domain := range validDomains {
		t.Run("valid_"+domain, func(t *testing.T) {
			err := errors.ValidateDomain(domain)
			if err != nil {
				t.Errorf("ValidateDomain(%q) = %v, want nil", domain, err)
			}
		})
	}

	invalidDomains := []struct {
		domain string
		reason string
	}{
		{"", "empty domain"},
		{".google.com", "starts with dot"},
		{"google.com.", "ends with dot"},
		{"-google.com", "starts with hyphen"},
		{"google.com-", "ends with hyphen"},
		{"google..com", "consecutive dots"},
		{"google@com", "invalid character"},
		{"google com", "space character"},
		{strings.Repeat("a", 254), "too long"},
	}

	for _, test := range invalidDomains {
		t.Run("invalid_"+test.reason, func(t *testing.T) {
			err := errors.ValidateDomain(test.domain)
			if err == nil {
				t.Errorf("ValidateDomain(%q) = nil, want error for %s", test.domain, test.reason)
			}
			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}
		})
	}
}

func TestValidateRecordType(t *testing.T) {
	validTypes := []string{"A", "AAAA", "MX", "CNAME", "TXT", "a", "aaaa", "mx", "cname", "txt"}
	for _, recordType := range validTypes {
		t.Run("valid_"+recordType, func(t *testing.T) {
			err := errors.ValidateRecordType(recordType)
			if err != nil {
				t.Errorf("ValidateRecordType(%q) = %v, want nil", recordType, err)
			}
		})
	}

	invalidTypes := []string{"", "INVALID", "NS", "SOA", "PTR", "123"}
	for _, recordType := range invalidTypes {
		t.Run("invalid_"+recordType, func(t *testing.T) {
			err := errors.ValidateRecordType(recordType)
			if err == nil {
				t.Errorf("ValidateRecordType(%q) = nil, want error", recordType)
			}
			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}
		})
	}
}

func TestValidateDNSServer(t *testing.T) {
	validServers := []string{
		"8.8.8.8",
		"1.1.1.1",
		"192.168.1.1",
		"::1",
		"2001:4860:4860::8888",
		"127.0.0.1",
	}

	for _, server := range validServers {
		t.Run("valid_"+server, func(t *testing.T) {
			err := errors.ValidateDNSServer(server)
			if err != nil {
				t.Errorf("ValidateDNSServer(%q) = %v, want nil", server, err)
			}
		})
	}

	invalidServers := []string{
		"",
		"invalid-ip",
		"256.256.256.256",
		"8.8.8",
		"8.8.8.8.8",
		"google.com",
		"127.0.0.2", // Other loopback addresses should warn
	}

	for _, server := range invalidServers {
		t.Run("invalid_"+server, func(t *testing.T) {
			err := errors.ValidateDNSServer(server)
			if err == nil {
				t.Errorf("ValidateDNSServer(%q) = nil, want error", server)
			}
			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}
		})
	}
}

func TestCLIParser_EdgeCases(t *testing.T) {
	parser := NewCLIParser()

	t.Run("flag without value", func(t *testing.T) {
		_, err := parser.Parse([]string{"-t"})
		if err == nil {
			t.Error("Parse() with flag without value should return error")
		}
	})

	t.Run("unknown flag", func(t *testing.T) {
		_, err := parser.Parse([]string{"-x", "value", "google.com"})
		if err == nil {
			t.Error("Parse() with unknown flag should return error")
		}
	})

	t.Run("mixed flag positions", func(t *testing.T) {
		// When domain comes first, flags after it are treated as additional arguments
		_, err := parser.Parse([]string{"google.com", "-t", "MX"})
		if err == nil {
			t.Error("Parse() with domain before flags should return error")
		}
		if !strings.Contains(err.Error(), "too many arguments") {
			t.Errorf("Parse() error = %v, want error containing 'too many arguments'", err)
		}
	})
}
