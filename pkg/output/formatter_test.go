package output

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go-dig/pkg/dns"
	"go-dig/pkg/errors"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Fatal("NewFormatter() returned nil")
	}
}

func TestFormatResult_Success(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{"93.184.216.34"},
		Server:     "8.8.8.8:53",
		QueryTime:  50 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	// Check that output contains expected elements with enhanced format
	expectedElements := []string{
		"; <<>> go-dig <<>> example.com A @8.8.8.8",
		";; Query time: 50 msec",
		";; SERVER: 8.8.8.8:53",
		";; WHEN:",
		";; ANSWER SECTION: (1 record)",
		"example.com                   \tIN\tA\t93.184.216.34",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_MultipleRecords(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "google.com",
		RecordType: "A",
		Records:    []string{"142.250.191.14", "142.250.191.46"},
		Server:     "1.1.1.1:53",
		QueryTime:  25 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	// Check that both records are present
	if !strings.Contains(output, "142.250.191.14") {
		t.Error("Expected output to contain first IP address")
	}
	if !strings.Contains(output, "142.250.191.46") {
		t.Error("Expected output to contain second IP address")
	}

	// Check for enhanced header format
	if !strings.Contains(output, ";; ANSWER SECTION: (2 records)") {
		t.Error("Expected enhanced answer section header with record count")
	}

	// Count the number of record lines with new format
	lines := strings.Split(output, "\n")
	recordCount := 0
	for _, line := range lines {
		if strings.Contains(line, "google.com                    \tIN\tA\t") {
			recordCount++
		}
	}

	if recordCount != 2 {
		t.Errorf("Expected 2 record lines, got %d", recordCount)
	}
}

func TestFormatResult_WithError(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "nonexistent.example",
		RecordType: "A",
		Records:    []string{},
		Server:     "8.8.8.8:53",
		QueryTime:  100 * time.Millisecond,
		Error:      errors.NewDNSError("domain 'nonexistent.example' not found (NXDOMAIN)", nil, "nonexistent.example", "8.8.8.8:53"),
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "Error: domain 'nonexistent.example' not found (NXDOMAIN)") {
		t.Errorf("Expected DNS error output, got '%s'", output)
	}
}

func TestFormatResult_NilResult(t *testing.T) {
	formatter := NewFormatter()

	output := formatter.FormatResult(nil)

	expected := "Error: no result to format\n"
	if output != expected {
		t.Errorf("Expected nil result error '%s', got '%s'", expected, output)
	}
}

func TestFormatError(t *testing.T) {
	formatter := NewFormatter()

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "simple error",
			err:      fmt.Errorf("connection failed"),
			expected: "Error: connection failed\n",
		},
		{
			name:     "input error",
			err:      errors.NewInputError("invalid domain format", nil),
			expected: "Error: invalid domain format\n",
		},
		{
			name:     "network error",
			err:      errors.NewNetworkError("DNS server timeout", nil, "8.8.8.8:53"),
			expected: "Error: DNS server timeout\n",
		},
		{
			name:     "DNS error",
			err:      errors.NewDNSError("domain not found", nil, "example.com", "8.8.8.8:53"),
			expected: "Error: domain not found\n",
		},
		{
			name:     "system error",
			err:      errors.NewSystemError("permission denied", nil),
			expected: "Error: permission denied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatter.FormatError(tt.err)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expected, output)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			expected: "500 usec",
		},
		{
			name:     "milliseconds",
			duration: 50 * time.Millisecond,
			expected: "50 msec",
		},
		{
			name:     "seconds as milliseconds",
			duration: 2 * time.Second,
			expected: "2000 msec",
		},
		{
			name:     "exactly 1 millisecond",
			duration: 1 * time.Millisecond,
			expected: "1 msec",
		},
		{
			name:     "less than 1 millisecond",
			duration: 999 * time.Microsecond,
			expected: "999 usec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatResult_ConsistentFormatting(t *testing.T) {
	formatter := NewFormatter()

	// Test with different domain lengths to ensure consistent formatting
	testCases := []struct {
		domain string
		ip     string
	}{
		{"a.com", "1.2.3.4"},
		{"verylongdomainname.example.com", "192.168.1.1"},
		{"short.co", "10.0.0.1"},
	}

	for _, tc := range testCases {
		result := &dns.Result{
			Domain:     tc.domain,
			RecordType: "A",
			Records:    []string{tc.ip},
			Server:     "8.8.8.8:53",
			QueryTime:  30 * time.Millisecond,
			Error:      nil,
		}

		output := formatter.FormatResult(result)

		// Check that the record line has consistent formatting with proper padding
		paddedDomain := fmt.Sprintf("%-30s", tc.domain)
		expectedRecordLine := paddedDomain + "\tIN\tA\t" + tc.ip
		if !strings.Contains(output, expectedRecordLine) {
			t.Errorf("Expected consistent formatting for domain '%s', missing: '%s'\nActual output:\n%s",
				tc.domain, expectedRecordLine, output)
		}
	}
}

// Additional tests for comprehensive error handling

func TestFormatDigError_InputError(t *testing.T) {
	formatter := NewFormatter()

	inputErr := errors.NewInputError("domain name cannot be empty", nil)
	inputErr.Domain = "invalid-domain"

	output := formatter.FormatError(inputErr)

	expectedElements := []string{
		"Error: domain name cannot be empty",
		"This is an input validation error",
		"Use 'go-dig --help' for usage information",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatDigError_NetworkError(t *testing.T) {
	formatter := NewFormatter()

	netErr := errors.NewNetworkError("DNS server timeout - server may be unreachable", nil, "8.8.8.8:53")

	output := formatter.FormatError(netErr)

	expectedElements := []string{
		"Error: DNS server timeout - server may be unreachable",
		"This is a network connectivity error",
		"DNS Server: 8.8.8.8:53",
		"Check your internet connection",
		"Try a different DNS server",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatDigError_DNSError(t *testing.T) {
	formatter := NewFormatter()

	dnsErr := errors.NewDNSError("domain 'nonexistent.example' not found (NXDOMAIN)", nil, "nonexistent.example", "1.1.1.1:53")

	output := formatter.FormatError(dnsErr)

	expectedElements := []string{
		"Error: domain 'nonexistent.example' not found (NXDOMAIN)",
		"This is a DNS resolution error",
		"Domain: nonexistent.example",
		"DNS Server: 1.1.1.1:53",
		"Verify the domain name is spelled correctly",
		"Check if the domain exists",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatDigError_SystemError(t *testing.T) {
	formatter := NewFormatter()

	sysErr := errors.NewSystemError("permission denied", fmt.Errorf("access denied"))

	output := formatter.FormatError(sysErr)

	expectedElements := []string{
		"Error: permission denied",
		"This is a system-level error",
		"Check if you have sufficient permissions",
		"Try running as administrator",
		"Underlying cause: access denied",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatDigError_WithCause(t *testing.T) {
	formatter := NewFormatter()

	cause := fmt.Errorf("connection timeout")
	netErr := errors.NewNetworkError("failed to connect to DNS server", cause, "8.8.8.8:53")

	output := formatter.FormatError(netErr)

	if !strings.Contains(output, "Underlying cause: connection timeout") {
		t.Errorf("Expected output to contain underlying cause, got: %s", output)
	}
}

func TestFormatDigError_WithoutOptionalFields(t *testing.T) {
	formatter := NewFormatter()

	// Test error without domain or server fields
	inputErr := errors.NewInputError("invalid command line arguments", nil)

	output := formatter.FormatError(inputErr)

	// Should not contain domain or server information
	if strings.Contains(output, "Domain:") {
		t.Errorf("Expected no domain information for generic input error, got: %s", output)
	}
	if strings.Contains(output, "DNS Server:") {
		t.Errorf("Expected no server information for generic input error, got: %s", output)
	}
}

func TestFormatResult_WithDigError(t *testing.T) {
	formatter := NewFormatter()

	// Test that FormatResult properly handles DigError in result
	dnsErr := errors.NewDNSError("no A records found", nil, "example.com", "8.8.8.8:53")
	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{},
		Server:     "8.8.8.8:53",
		QueryTime:  50 * time.Millisecond,
		Error:      dnsErr,
	}

	output := formatter.FormatResult(result)

	// Should use the enhanced error formatting
	if !strings.Contains(output, "This is a DNS resolution error") {
		t.Errorf("Expected enhanced DNS error formatting, got: %s", output)
	}
}

// Tests for enhanced formatting with different record types

func TestFormatResult_ARecord(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{"93.184.216.34"},
		Server:     "8.8.8.8:53",
		QueryTime:  50 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	// Check enhanced header format
	expectedElements := []string{
		"; <<>> go-dig <<>> example.com A @8.8.8.8",
		";; Query time: 50 msec",
		";; SERVER: 8.8.8.8:53",
		";; WHEN:",
		";; ANSWER SECTION: (1 record)",
		"example.com                   \tIN\tA\t93.184.216.34",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_AAAARecord(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "google.com",
		RecordType: "AAAA",
		Records:    []string{"2607:f8b0:4004:c1b::65"},
		Server:     "1.1.1.1:53",
		QueryTime:  30 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	expectedElements := []string{
		"; <<>> go-dig <<>> google.com AAAA @1.1.1.1",
		";; ANSWER SECTION: (1 record)",
		"google.com                    \tIN\tAAAA\t2607:f8b0:4004:c1b::65",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_MXRecord(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "MX",
		Records:    []string{"10 mail.example.com", "20 mail2.example.com"},
		Server:     "8.8.8.8:53",
		QueryTime:  75 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	expectedElements := []string{
		"; <<>> go-dig <<>> example.com MX @8.8.8.8",
		";; ANSWER SECTION: (2 records)",
		"example.com                   \tIN\tMX\t10 mail.example.com",
		"example.com                   \tIN\tMX\t20 mail2.example.com",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_CNAMERecord(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "www.example.com",
		RecordType: "CNAME",
		Records:    []string{"example.com"},
		Server:     "8.8.8.8:53",
		QueryTime:  25 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	expectedElements := []string{
		"; <<>> go-dig <<>> www.example.com CNAME @8.8.8.8",
		";; ANSWER SECTION: (1 record)",
		"www.example.com               \tIN\tCNAME\texample.com.",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_TXTRecord(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "TXT",
		Records:    []string{"v=spf1 include:_spf.google.com ~all", "simple-text"},
		Server:     "1.1.1.1:53",
		QueryTime:  40 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	expectedElements := []string{
		"; <<>> go-dig <<>> example.com TXT @1.1.1.1",
		";; ANSWER SECTION: (2 records)",
		"example.com                   \tIN\tTXT\t\"v=spf1 include:_spf.google.com ~all\"",
		"example.com                   \tIN\tTXT\tsimple-text",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_EmptyRecords(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{},
		Server:     "8.8.8.8:53",
		QueryTime:  100 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	expectedElements := []string{
		"; <<>> go-dig <<>> example.com A @8.8.8.8",
		";; ANSWER SECTION: (empty)",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nActual output:\n%s", element, output)
		}
	}
}

func TestFormatResult_NoServerSpecified(t *testing.T) {
	formatter := NewFormatter()

	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{"93.184.216.34"},
		Server:     "",
		QueryTime:  50 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)

	// Should not include @server in header when no server specified
	if strings.Contains(output, "@") {
		t.Errorf("Expected no server in header when server is empty, got: %s", output)
	}

	// But should still show server in metadata section
	if !strings.Contains(output, ";; SERVER:") {
		t.Errorf("Expected SERVER metadata line, got: %s", output)
	}
}

func TestFormatRecordValue(t *testing.T) {
	formatter := &formatter{}

	tests := []struct {
		name       string
		recordType string
		value      string
		expected   string
	}{
		{
			name:       "A record",
			recordType: "A",
			value:      "192.168.1.1",
			expected:   "192.168.1.1",
		},
		{
			name:       "AAAA record",
			recordType: "AAAA",
			value:      "2001:db8::1",
			expected:   "2001:db8::1",
		},
		{
			name:       "MX record",
			recordType: "MX",
			value:      "10 mail.example.com",
			expected:   "10 mail.example.com",
		},
		{
			name:       "CNAME record without dot",
			recordType: "CNAME",
			value:      "example.com",
			expected:   "example.com.",
		},
		{
			name:       "CNAME record with dot",
			recordType: "CNAME",
			value:      "example.com.",
			expected:   "example.com.",
		},
		{
			name:       "TXT record with spaces",
			recordType: "TXT",
			value:      "v=spf1 include:_spf.google.com ~all",
			expected:   "\"v=spf1 include:_spf.google.com ~all\"",
		},
		{
			name:       "TXT record without spaces",
			recordType: "TXT",
			value:      "simple-text",
			expected:   "simple-text",
		},
		{
			name:       "TXT record with quotes",
			recordType: "TXT",
			value:      "text with \"quotes\"",
			expected:   "\"text with \\\"quotes\\\"\"",
		},
		{
			name:       "Unknown record type",
			recordType: "NS",
			value:      "ns1.example.com",
			expected:   "ns1.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatRecordValue(tt.recordType, tt.value)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatResult_MultipleRecordsSingularPlural(t *testing.T) {
	formatter := NewFormatter()

	// Test singular
	result := &dns.Result{
		Domain:     "example.com",
		RecordType: "A",
		Records:    []string{"93.184.216.34"},
		Server:     "8.8.8.8:53",
		QueryTime:  50 * time.Millisecond,
		Error:      nil,
	}

	output := formatter.FormatResult(result)
	if !strings.Contains(output, "(1 record)") {
		t.Errorf("Expected singular 'record', got: %s", output)
	}

	// Test plural
	result.Records = []string{"93.184.216.34", "93.184.216.35"}
	output = formatter.FormatResult(result)
	if !strings.Contains(output, "(2 records)") {
		t.Errorf("Expected plural 'records', got: %s", output)
	}
}

func TestFormatResult_ConsistentDomainFormatting(t *testing.T) {
	formatter := NewFormatter()

	// Test with different domain lengths to ensure consistent formatting
	testCases := []struct {
		domain string
		ip     string
	}{
		{"a.co", "1.2.3.4"},
		{"verylongdomainname.example.com", "192.168.1.1"},
		{"medium-length.com", "10.0.0.1"},
	}

	for _, tc := range testCases {
		result := &dns.Result{
			Domain:     tc.domain,
			RecordType: "A",
			Records:    []string{tc.ip},
			Server:     "8.8.8.8:53",
			QueryTime:  30 * time.Millisecond,
			Error:      nil,
		}

		output := formatter.FormatResult(result)

		// Check that the record line has consistent formatting with proper padding
		// The padding should make the domain field 30 characters wide
		paddedDomain := fmt.Sprintf("%-30s", tc.domain)
		expectedRecordLine := paddedDomain + "\tIN\tA\t" + tc.ip

		if !strings.Contains(output, expectedRecordLine) {
			t.Errorf("Expected consistent formatting for domain '%s', missing: '%s'\nActual output:\n%s",
				tc.domain, expectedRecordLine, output)
		}
	}
}
