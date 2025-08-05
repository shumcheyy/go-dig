package output

import (
	"fmt"
	"strings"
	"time"

	"go-dig/pkg/dns"
	"go-dig/pkg/errors"
)

// Formatter interface defines output formatting functionality
type Formatter interface {
	FormatResult(result *dns.Result) string
	FormatError(err error) string
}

// formatter implements the Formatter interface
type formatter struct{}

// NewFormatter creates a new output formatter
func NewFormatter() Formatter {
	return &formatter{}
}

// FormatResult formats a successful DNS query result for display
func (f *formatter) FormatResult(result *dns.Result) string {
	if result == nil {
		return f.FormatError(fmt.Errorf("no result to format"))
	}

	// Handle error results
	if result.Error != nil {
		return f.FormatError(result.Error)
	}

	var output strings.Builder

	// Header with query information - show what was queried and where
	output.WriteString(fmt.Sprintf("; <<>> go-dig <<>> %s %s", result.Domain, result.RecordType))
	if result.Server != "" {
		output.WriteString(fmt.Sprintf(" @%s", strings.TrimSuffix(result.Server, ":53")))
	}
	output.WriteString("\n")

	// Query metadata
	output.WriteString(fmt.Sprintf(";; Query time: %v\n", formatDuration(result.QueryTime)))
	output.WriteString(fmt.Sprintf(";; SERVER: %s\n", result.Server))
	output.WriteString(fmt.Sprintf(";; WHEN: %s\n", time.Now().Format("Mon Jan 02 15:04:05 MST 2006")))
	output.WriteString("\n")

	// Answer section with record count
	recordCount := len(result.Records)
	if recordCount == 0 {
		output.WriteString(";; ANSWER SECTION: (empty)\n")
	} else {
		output.WriteString(fmt.Sprintf(";; ANSWER SECTION: (%d record", recordCount))
		if recordCount != 1 {
			output.WriteString("s")
		}
		output.WriteString(")\n")

		// Format records based on type
		for _, record := range result.Records {
			formattedRecord := f.formatRecordValue(result.RecordType, record)
			output.WriteString(fmt.Sprintf("%-30s\tIN\t%s\t%s\n",
				result.Domain, result.RecordType, formattedRecord))
		}
	}

	return output.String()
}

// FormatError formats error messages for display with enhanced error handling
func (f *formatter) FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's a DigError for enhanced formatting
	if digErr, ok := err.(*errors.DigError); ok {
		return f.formatDigError(digErr)
	}

	// Fallback for other error types
	return fmt.Sprintf("Error: %s\n", err.Error())
}

// formatDigError formats DigError with context-specific information
func (f *formatter) formatDigError(digErr *errors.DigError) string {
	var output strings.Builder

	// Write the main error message
	output.WriteString(fmt.Sprintf("Error: %s\n", digErr.Message))

	// Add context-specific information based on error type
	switch digErr.Type {
	case errors.ErrorTypeInput:
		output.WriteString("\nThis is an input validation error. Please check your command line arguments.\n")
		if digErr.Domain != "" {
			output.WriteString(fmt.Sprintf("Domain: %s\n", digErr.Domain))
		}
		output.WriteString("Use 'go-dig --help' for usage information.\n")

	case errors.ErrorTypeNetwork:
		output.WriteString("\nThis is a network connectivity error.\n")
		if digErr.Server != "" {
			output.WriteString(fmt.Sprintf("DNS Server: %s\n", digErr.Server))
		}
		output.WriteString("Troubleshooting suggestions:\n")
		output.WriteString("- Check your internet connection\n")
		output.WriteString("- Try a different DNS server (e.g., -s 8.8.8.8)\n")
		output.WriteString("- Verify the DNS server IP address is correct\n")

	case errors.ErrorTypeDNS:
		output.WriteString("\nThis is a DNS resolution error.\n")
		if digErr.Domain != "" {
			output.WriteString(fmt.Sprintf("Domain: %s\n", digErr.Domain))
		}
		if digErr.Server != "" {
			output.WriteString(fmt.Sprintf("DNS Server: %s\n", digErr.Server))
		}
		output.WriteString("Troubleshooting suggestions:\n")
		output.WriteString("- Verify the domain name is spelled correctly\n")
		output.WriteString("- Check if the domain exists\n")
		output.WriteString("- Try querying a different record type\n")
		output.WriteString("- Try a different DNS server\n")

	case errors.ErrorTypeSystem:
		output.WriteString("\nThis is a system-level error.\n")
		output.WriteString("Troubleshooting suggestions:\n")
		output.WriteString("- Check if you have sufficient permissions\n")
		output.WriteString("- Verify system DNS configuration\n")
		output.WriteString("- Try running as administrator if needed\n")
	}

	// Add underlying cause if available
	if digErr.Cause != nil {
		output.WriteString(fmt.Sprintf("\nUnderlying cause: %v\n", digErr.Cause))
	}

	return output.String()
}

// formatRecordValue formats a record value based on its type for proper display
func (f *formatter) formatRecordValue(recordType, value string) string {
	switch strings.ToUpper(recordType) {
	case "A", "AAAA":
		// IP addresses are displayed as-is
		return value
	case "MX":
		// MX records already include priority from DNS client
		return value
	case "CNAME":
		// CNAME records should end with a dot if they don't already
		if !strings.HasSuffix(value, ".") {
			return value + "."
		}
		return value
	case "TXT":
		// TXT records should be quoted if they contain spaces or special characters
		if strings.Contains(value, " ") || strings.ContainsAny(value, "\"'\\") {
			// Escape existing quotes and wrap in quotes
			escaped := strings.ReplaceAll(value, "\"", "\\\"")
			return fmt.Sprintf("\"%s\"", escaped)
		}
		return value
	default:
		// For unknown types, return as-is
		return value
	}
}

// formatDuration formats duration in a human-readable way similar to dig
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%d usec", d.Microseconds())
	}
	return fmt.Sprintf("%d msec", d.Milliseconds())
}
