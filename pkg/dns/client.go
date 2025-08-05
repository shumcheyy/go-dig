package dns

import (
	"fmt"
	"go-dig/pkg/errors"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Result holds the results of a DNS query including timing information
type Result struct {
	Domain     string
	RecordType string
	Records    []string
	Server     string
	QueryTime  time.Duration
	Error      error
}

// Client interface defines the DNS query functionality
type Client interface {
	Query(domain, recordType, server string) (*Result, error)
	SetTimeout(duration time.Duration)
}

// client implements the Client interface
type client struct {
	timeout time.Duration
}

// NewClient creates a new DNS client with default timeout
func NewClient() Client {
	return &client{
		timeout: 5 * time.Second, // Default 5 second timeout
	}
}

// SetTimeout sets the query timeout duration
func (c *client) SetTimeout(duration time.Duration) {
	c.timeout = duration
}

// Query performs a DNS query for the specified domain and record type
func (c *client) Query(domain, recordType, server string) (*Result, error) {
	result := &Result{
		Domain:     domain,
		RecordType: recordType,
		Server:     server,
		Records:    []string{},
	}

	// Validate domain
	if err := errors.ValidateDomain(domain); err != nil {
		result.Error = err
		return result, err
	}

	// Validate record type
	if err := errors.ValidateRecordType(recordType); err != nil {
		result.Error = err
		return result, err
	}

	// Prepare DNS server
	var finalServer string
	if server == "" {
		// Use system default DNS server
		systemDNS, err := getSystemDNS()
		if err != nil {
			// Fallback to Google DNS if system DNS cannot be determined
			finalServer = "8.8.8.8:53"
		} else {
			finalServer = systemDNS
		}
	} else {
		// Check if server already has port
		// For IPv6 addresses, we need to be more careful about detection
		if strings.HasPrefix(server, "[") && strings.Contains(server, "]:") {
			// IPv6 with port in brackets format [::1]:53
			host, port, err := net.SplitHostPort(server)
			if err != nil {
				err := errors.NewInputError(fmt.Sprintf("invalid DNS server format: %s", server), err)
				result.Error = err
				return result, err
			}
			if err := errors.ValidateDNSServer(host); err != nil {
				result.Error = err
				return result, err
			}
			if err := errors.ValidateDNSPort(port); err != nil {
				result.Error = err
				return result, err
			}
			finalServer = server
		} else if !strings.Contains(server, ":") || net.ParseIP(server) != nil {
			// IPv4 address or IPv6 address without port
			if err := errors.ValidateDNSServer(server); err != nil {
				result.Error = err
				return result, err
			}
			// Handle IPv6 addresses properly
			if strings.Contains(server, ":") && net.ParseIP(server) != nil {
				// This is an IPv6 address without brackets
				finalServer = "[" + server + "]:53"
			} else {
				finalServer = server + ":53"
			}
		} else {
			// IPv4 with port
			host, port, err := net.SplitHostPort(server)
			if err != nil {
				err := errors.NewInputError(fmt.Sprintf("invalid DNS server format: %s", server), err)
				result.Error = err
				return result, err
			}
			if err := errors.ValidateDNSServer(host); err != nil {
				result.Error = err
				return result, err
			}
			if err := errors.ValidateDNSPort(port); err != nil {
				result.Error = err
				return result, err
			}
			finalServer = server
		}
	}
	result.Server = finalServer

	// Create DNS client
	dnsClient := new(dns.Client)
	dnsClient.Timeout = c.timeout

	// Determine DNS query type
	var queryType uint16
	recordTypeUpper := strings.ToUpper(recordType)
	switch recordTypeUpper {
	case "A":
		queryType = dns.TypeA
	case "AAAA":
		queryType = dns.TypeAAAA
	case "MX":
		queryType = dns.TypeMX
	case "CNAME":
		queryType = dns.TypeCNAME
	case "TXT":
		queryType = dns.TypeTXT
	default:
		// This should not happen due to validation, but handle it gracefully
		err := errors.NewInputError(fmt.Sprintf("record type '%s' not supported", recordType), nil)
		result.Error = err
		return result, err
	}

	// Create DNS message
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), queryType)
	msg.RecursionDesired = true

	// Perform the query and measure time
	startTime := time.Now()
	response, _, err := dnsClient.Exchange(msg, finalServer)
	result.QueryTime = time.Since(startTime)

	if err != nil {
		// Classify and wrap the network error
		netErr := errors.ClassifyNetworkError(err, finalServer)
		result.Error = netErr
		return result, netErr
	}

	if response == nil {
		err := errors.NewNetworkError("no response received from DNS server", nil, finalServer)
		result.Error = err
		return result, err
	}

	// Check response code and create appropriate DNS errors
	if response.Rcode != dns.RcodeSuccess {
		var dnsErr *errors.DigError
		switch response.Rcode {
		case dns.RcodeNameError:
			dnsErr = errors.NewDNSError(fmt.Sprintf("domain '%s' not found (NXDOMAIN)", domain), nil, domain, finalServer)
		case dns.RcodeServerFailure:
			dnsErr = errors.NewDNSError("DNS server experienced an internal failure", nil, domain, finalServer)
		case dns.RcodeRefused:
			dnsErr = errors.NewDNSError("DNS server refused the query", nil, domain, finalServer)
		case dns.RcodeNotImplemented:
			dnsErr = errors.NewDNSError("DNS server does not support this query type", nil, domain, finalServer)
		case dns.RcodeFormatError:
			dnsErr = errors.NewDNSError("DNS query format error", nil, domain, finalServer)
		default:
			dnsErr = errors.NewDNSError(fmt.Sprintf("DNS query failed with response code %d", response.Rcode), nil, domain, finalServer)
		}
		result.Error = dnsErr
		return result, dnsErr
	}

	// Extract records from response based on record type
	for _, answer := range response.Answer {
		switch recordTypeUpper {
		case "A":
			if aRecord, ok := answer.(*dns.A); ok {
				result.Records = append(result.Records, aRecord.A.String())
			}
		case "AAAA":
			if aaaaRecord, ok := answer.(*dns.AAAA); ok {
				result.Records = append(result.Records, aaaaRecord.AAAA.String())
			}
		case "MX":
			if mxRecord, ok := answer.(*dns.MX); ok {
				result.Records = append(result.Records, fmt.Sprintf("%d %s", mxRecord.Preference, mxRecord.Mx))
			}
		case "CNAME":
			if cnameRecord, ok := answer.(*dns.CNAME); ok {
				result.Records = append(result.Records, cnameRecord.Target)
			}
		case "TXT":
			if txtRecord, ok := answer.(*dns.TXT); ok {
				// TXT records can have multiple strings, join them
				result.Records = append(result.Records, strings.Join(txtRecord.Txt, " "))
			}
		}
	}

	if len(result.Records) == 0 {
		err := errors.NewDNSError(fmt.Sprintf("no %s records found for domain '%s'", recordTypeUpper, domain), nil, domain, finalServer)
		result.Error = err
		return result, err
	}

	return result, nil
}

// getSystemDNS attempts to determine the system's default DNS server
func getSystemDNS() (string, error) {
	// Try to get system DNS configuration
	if runtime.GOOS == "windows" {
		return getWindowsDNS()
	}

	// For non-Windows systems, try to read /etc/resolv.conf
	return getUnixDNS()
}

// getWindowsDNS attempts to get DNS server from Windows system
func getWindowsDNS() (string, error) {
	// On Windows, try to use the miekg/dns library's system detection
	// This will attempt to read Windows DNS configuration
	config, err := dns.ClientConfigFromFile("")
	if err == nil && len(config.Servers) > 0 {
		return config.Servers[0] + ":53", nil
	}

	// If that fails, return an error to trigger fallback to Google DNS
	return "", fmt.Errorf("could not determine Windows DNS server configuration")
}

// getUnixDNS attempts to get DNS server from Unix-like systems
func getUnixDNS() (string, error) {
	// Try to read /etc/resolv.conf
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return "", fmt.Errorf("could not read resolv.conf: %v", err)
	}

	if len(config.Servers) == 0 {
		return "", fmt.Errorf("no DNS servers found in resolv.conf")
	}

	// Return the first DNS server with port
	return config.Servers[0] + ":53", nil
}
