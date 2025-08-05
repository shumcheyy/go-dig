package errors

import (
	"fmt"
	"net"
	"strings"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	ErrorTypeInput ErrorType = iota
	ErrorTypeNetwork
	ErrorTypeDNS
	ErrorTypeSystem
)

// String returns a string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeInput:
		return "Input"
	case ErrorTypeNetwork:
		return "Network"
	case ErrorTypeDNS:
		return "DNS"
	case ErrorTypeSystem:
		return "System"
	default:
		return "Unknown"
	}
}

// DigError represents a structured error with type and context
type DigError struct {
	Type    ErrorType
	Message string
	Cause   error
	Domain  string
	Server  string
}

// Error implements the error interface
func (de *DigError) Error() string {
	if de.Cause != nil {
		return fmt.Sprintf("%s error: %s (caused by: %v)", de.Type.String(), de.Message, de.Cause)
	}
	return fmt.Sprintf("%s error: %s", de.Type.String(), de.Message)
}

// Unwrap returns the underlying cause error
func (de *DigError) Unwrap() error {
	return de.Cause
}

// NewInputError creates a new input validation error
func NewInputError(message string, cause error) *DigError {
	return &DigError{
		Type:    ErrorTypeInput,
		Message: message,
		Cause:   cause,
	}
}

// NewNetworkError creates a new network-related error
func NewNetworkError(message string, cause error, server string) *DigError {
	return &DigError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Cause:   cause,
		Server:  server,
	}
}

// NewDNSError creates a new DNS-related error
func NewDNSError(message string, cause error, domain, server string) *DigError {
	return &DigError{
		Type:    ErrorTypeDNS,
		Message: message,
		Cause:   cause,
		Domain:  domain,
		Server:  server,
	}
}

// NewSystemError creates a new system-level error
func NewSystemError(message string, cause error) *DigError {
	return &DigError{
		Type:    ErrorTypeSystem,
		Message: message,
		Cause:   cause,
	}
}

// IsInputError checks if the error is an input validation error
func IsInputError(err error) bool {
	if digErr, ok := err.(*DigError); ok {
		return digErr.Type == ErrorTypeInput
	}
	return false
}

// IsNetworkError checks if the error is a network-related error
func IsNetworkError(err error) bool {
	if digErr, ok := err.(*DigError); ok {
		return digErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsDNSError checks if the error is a DNS-related error
func IsDNSError(err error) bool {
	if digErr, ok := err.(*DigError); ok {
		return digErr.Type == ErrorTypeDNS
	}
	return false
}

// IsSystemError checks if the error is a system-level error
func IsSystemError(err error) bool {
	if digErr, ok := err.(*DigError); ok {
		return digErr.Type == ErrorTypeSystem
	}
	return false
}

// ClassifyNetworkError analyzes a network error and returns appropriate DigError
func ClassifyNetworkError(err error, server string) *DigError {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check for common network error patterns
	switch {
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"):
		return NewNetworkError("DNS server timeout - server may be unreachable or overloaded", err, server)
	case strings.Contains(errStr, "connection refused"):
		return NewNetworkError("DNS server refused connection - server may be down or not accepting queries", err, server)
	case strings.Contains(errStr, "no such host"):
		return NewNetworkError("DNS server hostname could not be resolved", err, server)
	case strings.Contains(errStr, "network unreachable"):
		return NewNetworkError("Network unreachable - check your internet connection", err, server)
	case strings.Contains(errStr, "permission denied"):
		return NewNetworkError("Permission denied - may need elevated privileges", err, server)
	default:
		return NewNetworkError("Network error occurred while contacting DNS server", err, server)
	}
}

// ValidateDomain performs comprehensive domain name validation
func ValidateDomain(domain string) error {
	if domain == "" {
		return NewInputError("domain name cannot be empty", nil)
	}

	// Check domain length (max 253 characters per RFC)
	if len(domain) > 253 {
		return NewInputError(fmt.Sprintf("domain name too long (%d characters, max 253)", len(domain)), nil)
	}

	// Check for invalid characters (allow underscores for DNS records like _dmarc)
	for i, r := range domain {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_') {
			return NewInputError(fmt.Sprintf("domain contains invalid character '%c' at position %d", r, i+1), nil)
		}
	}

	// Check for consecutive dots
	if strings.Contains(domain, "..") {
		return NewInputError("domain cannot contain consecutive dots", nil)
	}

	// Check if domain starts or ends with invalid characters
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return NewInputError("domain cannot start or end with a dot", nil)
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return NewInputError("domain cannot start or end with a hyphen", nil)
	}
	if strings.HasSuffix(domain, "_") {
		return NewInputError("domain cannot end with an underscore", nil)
	}

	// Check each label (part between dots)
	labels := strings.Split(domain, ".")
	for i, label := range labels {
		if len(label) == 0 {
			return NewInputError("domain cannot have empty labels", nil)
		}
		if len(label) > 63 {
			return NewInputError(fmt.Sprintf("domain label '%s' too long (%d characters, max 63)", label, len(label)), nil)
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return NewInputError(fmt.Sprintf("domain label '%s' cannot start or end with hyphen", label), nil)
		}
		if strings.HasSuffix(label, "_") {
			return NewInputError(fmt.Sprintf("domain label '%s' cannot end with underscore", label), nil)
		}

		// Check if it's the last label (TLD) and ensure it's not all numeric
		if i == len(labels)-1 {
			allNumeric := true
			for _, r := range label {
				if r < '0' || r > '9' {
					allNumeric = false
					break
				}
			}
			if allNumeric {
				return NewInputError("top-level domain cannot be all numeric", nil)
			}
		}
	}

	return nil
}

// ValidateDNSServer validates DNS server IP address format
func ValidateDNSServer(server string) error {
	if server == "" {
		return NewInputError("DNS server cannot be empty", nil)
	}

	// Parse as IP address
	ip := net.ParseIP(server)
	if ip == nil {
		return NewInputError(fmt.Sprintf("'%s' is not a valid IP address", server), nil)
	}

	// Additional validation for reserved/special addresses
	if ip.IsLoopback() && !ip.Equal(net.IPv4(127, 0, 0, 1)) && !ip.Equal(net.IPv6loopback) {
		return NewInputError("loopback addresses other than 127.0.0.1 or ::1 are not recommended for DNS", nil)
	}

	return nil
}

// ValidateRecordType validates DNS record type
func ValidateRecordType(recordType string) error {
	if recordType == "" {
		return NewInputError("record type cannot be empty", nil)
	}

	validTypes := map[string]bool{
		"A":     true,
		"AAAA":  true,
		"MX":    true,
		"CNAME": true,
		"TXT":   true,
	}

	recordType = strings.ToUpper(recordType)
	if !validTypes[recordType] {
		return NewInputError(fmt.Sprintf("unsupported record type '%s' (supported: A, AAAA, MX, CNAME, TXT)", recordType), nil)
	}

	return nil
}

// ValidateDNSPort validates DNS server port
func ValidateDNSPort(port string) error {
	if port == "" {
		return NewInputError("DNS server port cannot be empty", nil)
	}

	// Parse port as integer
	portNum := 0
	for _, r := range port {
		if r < '0' || r > '9' {
			return NewInputError(fmt.Sprintf("DNS server port '%s' must be numeric", port), nil)
		}
		portNum = portNum*10 + int(r-'0')
		if portNum > 65535 {
			return NewInputError(fmt.Sprintf("DNS server port '%s' is out of range (1-65535)", port), nil)
		}
	}

	if portNum < 1 || portNum > 65535 {
		return NewInputError(fmt.Sprintf("DNS server port '%s' is out of range (1-65535)", port), nil)
	}

	return nil
}
