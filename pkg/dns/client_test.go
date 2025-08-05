package dns

import (
	"go-dig/pkg/errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// mockDNSServer creates a mock DNS server for testing
func mockDNSServer(t *testing.T, handler func(w dns.ResponseWriter, r *dns.Msg)) (string, func()) {
	// Create a UDP listener on a random port
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen on UDP: %v", err)
	}

	server := &dns.Server{
		PacketConn: conn,
		Handler:    dns.HandlerFunc(handler),
	}

	go func() {
		if err := server.ActivateAndServe(); err != nil {
			t.Logf("Mock DNS server error: %v", err)
		}
	}()

	// Return server address and cleanup function
	return conn.LocalAddr().String(), func() {
		server.Shutdown()
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	// Test that client implements Client interface
	var _ Client = client
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient()
	timeout := 10 * time.Second
	client.SetTimeout(timeout)

	// We can't directly test the timeout value since it's private,
	// but we can test that the method doesn't panic
}

func TestClient_Query_EmptyDomain(t *testing.T) {
	client := NewClient()
	result, err := client.Query("", "A", "8.8.8.8")

	if err == nil {
		t.Fatal("Expected error for empty domain")
	}

	if !errors.IsInputError(err) {
		t.Errorf("Expected input error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if result.Domain != "" {
		t.Errorf("Expected empty domain, got %s", result.Domain)
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected error message about empty domain, got: %s", err.Error())
	}
}

func TestClient_Query_UnsupportedRecordType(t *testing.T) {
	client := NewClient()
	result, err := client.Query("example.com", "INVALID", "8.8.8.8")

	if err == nil {
		t.Fatal("Expected error for unsupported record type")
	}

	if !errors.IsInputError(err) {
		t.Errorf("Expected input error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "unsupported record type") {
		t.Errorf("Expected error about unsupported record type, got: %s", err.Error())
	}
}

func TestClient_Query_InvalidDNSServer(t *testing.T) {
	client := NewClient()
	result, err := client.Query("example.com", "A", "invalid-server")

	if err == nil {
		t.Fatal("Expected error for invalid DNS server")
	}

	if !errors.IsInputError(err) {
		t.Errorf("Expected input error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "not a valid IP address") {
		t.Errorf("Expected error about invalid IP, got: %s", err.Error())
	}
}

func TestClient_Query_SuccessfulARecord(t *testing.T) {
	// Create mock DNS server that returns A record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add A record to response
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("192.0.2.1"),
		}
		msg.Answer = append(msg.Answer, aRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %s", result.Domain)
	}

	if result.RecordType != "A" {
		t.Errorf("Expected record type 'A', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "192.0.2.1" {
		t.Errorf("Expected IP '192.0.2.1', got %s", result.Records[0])
	}

	if result.QueryTime <= 0 {
		t.Error("Expected positive query time")
	}

	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_MultipleARecords(t *testing.T) {
	// Create mock DNS server that returns multiple A records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add multiple A records to response
		ips := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
		for _, ip := range ips {
			aRecord := &dns.A{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				A: net.ParseIP(ip),
			}
			msg.Answer = append(msg.Answer, aRecord)
		}

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	expectedIPs := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
	if len(result.Records) != len(expectedIPs) {
		t.Fatalf("Expected %d records, got %d", len(expectedIPs), len(result.Records))
	}

	for i, expectedIP := range expectedIPs {
		if result.Records[i] != expectedIP {
			t.Errorf("Expected IP %s at index %d, got %s", expectedIP, i, result.Records[i])
		}
	}
}

func TestClient_Query_NXDOMAIN(t *testing.T) {
	// Create mock DNS server that returns NXDOMAIN
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeNameError // NXDOMAIN

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("nonexistent.example.com", "A", serverAddr)

	if err == nil {
		t.Fatal("Expected error for NXDOMAIN response")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "not found (NXDOMAIN)") {
		t.Errorf("Expected NXDOMAIN error, got: %s", err.Error())
	}
}

func TestClient_Query_NoARecords(t *testing.T) {
	// Create mock DNS server that returns successful response but no A records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		// No answer records added

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err == nil {
		t.Fatal("Expected error for no A records")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "no A records found") {
		t.Errorf("Expected error about no A records, got: %s", err.Error())
	}
}

func TestClient_Query_DefaultServer(t *testing.T) {
	client := NewClient()
	result, _ := client.Query("example.com", "A", "")

	// Should use default server (8.8.8.8:53)
	if result.Server != "8.8.8.8:53" {
		t.Errorf("Expected default server '8.8.8.8:53', got %s", result.Server)
	}
}

func TestClient_Query_ServerWithPort(t *testing.T) {
	client := NewClient()
	result, _ := client.Query("example.com", "A", "1.1.1.1:53")

	// Should preserve server with port
	if result.Server != "1.1.1.1:53" {
		t.Errorf("Expected server '1.1.1.1:53', got %s", result.Server)
	}
}

func TestClient_Query_ServerWithoutPort(t *testing.T) {
	client := NewClient()
	result, _ := client.Query("example.com", "A", "1.1.1.1")

	// Should add default port
	if result.Server != "1.1.1.1:53" {
		t.Errorf("Expected server '1.1.1.1:53', got %s", result.Server)
	}
}

func TestClient_Query_Timeout(t *testing.T) {
	// Create mock DNS server that doesn't respond
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Don't respond to simulate timeout
		time.Sleep(2 * time.Second)
	})
	defer cleanup()

	client := NewClient()
	client.SetTimeout(100 * time.Millisecond) // Very short timeout

	start := time.Now()
	result, err := client.Query("example.com", "A", serverAddr)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	// Should timeout quickly
	if elapsed > 500*time.Millisecond {
		t.Errorf("Query took too long: %v", elapsed)
	}
}

// Additional tests for comprehensive error handling

func TestClient_Query_InvalidDomain(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name   string
		domain string
	}{
		{"domain with invalid chars", "example@.com"},
		{"domain too long", strings.Repeat("a", 254)},
		{"consecutive dots", "example..com"},
		{"starts with dot", ".example.com"},
		{"ends with dot", "example.com."},
		{"starts with hyphen", "-example.com"},
		{"ends with hyphen", "example.com-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Query(tt.domain, "A", "8.8.8.8")

			if err == nil {
				t.Fatal("Expected error for invalid domain")
			}

			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}

			if result.Error == nil {
				t.Fatal("Expected result.Error to be set")
			}
		})
	}
}

func TestClient_Query_InvalidRecordType(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name       string
		recordType string
	}{
		{"empty record type", ""},
		{"invalid record type", "INVALID"},
		{"numeric record type", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Query("example.com", tt.recordType, "8.8.8.8")

			if err == nil {
				t.Fatal("Expected error for invalid record type")
			}

			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T", err)
			}

			if result.Error == nil {
				t.Fatal("Expected result.Error to be set")
			}
		})
	}
}

func TestClient_Query_DNSServerFailure(t *testing.T) {
	// Create mock DNS server that returns server failure
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeServerFailure

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err == nil {
		t.Fatal("Expected error for server failure")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "internal failure") {
		t.Errorf("Expected server failure error, got: %s", err.Error())
	}
}

func TestClient_Query_DNSRefused(t *testing.T) {
	// Create mock DNS server that refuses the query
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Rcode = dns.RcodeRefused

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err == nil {
		t.Fatal("Expected error for refused query")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "refused") {
		t.Errorf("Expected refused error, got: %s", err.Error())
	}
}

func TestClient_Query_NetworkTimeout(t *testing.T) {
	// Create mock DNS server that doesn't respond to simulate network timeout
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Don't respond to simulate timeout
		time.Sleep(2 * time.Second)
	})
	defer cleanup()

	client := NewClient()
	client.SetTimeout(100 * time.Millisecond) // Very short timeout

	result, err := client.Query("example.com", "A", serverAddr)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !errors.IsNetworkError(err) {
		t.Errorf("Expected network error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	// Should be classified as a timeout error
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %s", err.Error())
	}
}

func TestClient_Query_ValidRecordTypes(t *testing.T) {
	client := NewClient()

	// Test that all valid record types pass validation
	validTypes := []string{"A", "AAAA", "MX", "CNAME", "TXT", "a", "aaaa", "mx", "cname", "txt"}

	for _, recordType := range validTypes {
		t.Run("record_type_"+recordType, func(t *testing.T) {
			_, err := client.Query("example.com", recordType, "8.8.8.8")

			// All record types should now be supported, so we should not get input validation errors
			// We might get network errors (which is fine for this test)
			if err != nil && errors.IsInputError(err) && strings.Contains(err.Error(), "unsupported record type") {
				t.Errorf("Record type '%s' should be supported, got: %s", recordType, err.Error())
			}
		})
	}
}

// Tests for AAAA record support

func TestClient_Query_SuccessfulAAAARecord(t *testing.T) {
	// Create mock DNS server that returns AAAA record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add AAAA record to response
		aaaaRecord := &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			AAAA: net.ParseIP("2001:db8::1"),
		}
		msg.Answer = append(msg.Answer, aaaaRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "AAAA", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %s", result.Domain)
	}

	if result.RecordType != "AAAA" {
		t.Errorf("Expected record type 'AAAA', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "2001:db8::1" {
		t.Errorf("Expected IPv6 '2001:db8::1', got %s", result.Records[0])
	}

	if result.QueryTime <= 0 {
		t.Error("Expected positive query time")
	}

	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_MultipleAAAARecords(t *testing.T) {
	// Create mock DNS server that returns multiple AAAA records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add multiple AAAA records to response
		ipv6s := []string{"2001:db8::1", "2001:db8::2", "2001:db8::3"}
		for _, ipv6 := range ipv6s {
			aaaaRecord := &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				AAAA: net.ParseIP(ipv6),
			}
			msg.Answer = append(msg.Answer, aaaaRecord)
		}

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "AAAA", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	expectedIPv6s := []string{"2001:db8::1", "2001:db8::2", "2001:db8::3"}
	if len(result.Records) != len(expectedIPv6s) {
		t.Fatalf("Expected %d records, got %d", len(expectedIPv6s), len(result.Records))
	}

	for i, expectedIPv6 := range expectedIPv6s {
		if result.Records[i] != expectedIPv6 {
			t.Errorf("Expected IPv6 %s at index %d, got %s", expectedIPv6, i, result.Records[i])
		}
	}
}

func TestClient_Query_NoAAAARecords(t *testing.T) {
	// Create mock DNS server that returns successful response but no AAAA records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		// No answer records added

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "AAAA", serverAddr)

	if err == nil {
		t.Fatal("Expected error for no AAAA records")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "no AAAA records found") {
		t.Errorf("Expected error about no AAAA records, got: %s", err.Error())
	}
}

// Tests for MX record support

func TestClient_Query_SuccessfulMXRecord(t *testing.T) {
	// Create mock DNS server that returns MX record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Add small delay to ensure query time is measurable
		time.Sleep(1 * time.Millisecond)

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add MX record to response
		mxRecord := &dns.MX{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeMX,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Preference: 10,
			Mx:         "mail.example.com.",
		}
		msg.Answer = append(msg.Answer, mxRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "MX", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %s", result.Domain)
	}

	if result.RecordType != "MX" {
		t.Errorf("Expected record type 'MX', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "10 mail.example.com." {
		t.Errorf("Expected MX '10 mail.example.com.', got %s", result.Records[0])
	}

	if result.QueryTime <= 0 {
		t.Error("Expected positive query time")
	}

	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_MultipleMXRecords(t *testing.T) {
	// Create mock DNS server that returns multiple MX records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add multiple MX records to response
		mxRecords := []struct {
			preference uint16
			mx         string
		}{
			{10, "mail1.example.com."},
			{20, "mail2.example.com."},
			{30, "mail3.example.com."},
		}

		for _, mx := range mxRecords {
			mxRecord := &dns.MX{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Preference: mx.preference,
				Mx:         mx.mx,
			}
			msg.Answer = append(msg.Answer, mxRecord)
		}

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "MX", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	expectedMXs := []string{"10 mail1.example.com.", "20 mail2.example.com.", "30 mail3.example.com."}
	if len(result.Records) != len(expectedMXs) {
		t.Fatalf("Expected %d records, got %d", len(expectedMXs), len(result.Records))
	}

	for i, expectedMX := range expectedMXs {
		if result.Records[i] != expectedMX {
			t.Errorf("Expected MX %s at index %d, got %s", expectedMX, i, result.Records[i])
		}
	}
}

func TestClient_Query_NoMXRecords(t *testing.T) {
	// Create mock DNS server that returns successful response but no MX records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		// No answer records added

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "MX", serverAddr)

	if err == nil {
		t.Fatal("Expected error for no MX records")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "no MX records found") {
		t.Errorf("Expected error about no MX records, got: %s", err.Error())
	}
}

// Tests for CNAME record support

func TestClient_Query_SuccessfulCNAMERecord(t *testing.T) {
	// Create mock DNS server that returns CNAME record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Add small delay to ensure query time is measurable
		time.Sleep(1 * time.Millisecond)

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add CNAME record to response
		cnameRecord := &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Target: "target.example.com.",
		}
		msg.Answer = append(msg.Answer, cnameRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("alias.example.com", "CNAME", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.Domain != "alias.example.com" {
		t.Errorf("Expected domain 'alias.example.com', got %s", result.Domain)
	}

	if result.RecordType != "CNAME" {
		t.Errorf("Expected record type 'CNAME', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "target.example.com." {
		t.Errorf("Expected CNAME 'target.example.com.', got %s", result.Records[0])
	}

	if result.QueryTime <= 0 {
		t.Error("Expected positive query time")
	}

	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_MultipleCNAMERecords(t *testing.T) {
	// Create mock DNS server that returns multiple CNAME records (unusual but possible)
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add multiple CNAME records to response
		targets := []string{"target1.example.com.", "target2.example.com."}
		for _, target := range targets {
			cnameRecord := &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Target: target,
			}
			msg.Answer = append(msg.Answer, cnameRecord)
		}

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("alias.example.com", "CNAME", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	expectedTargets := []string{"target1.example.com.", "target2.example.com."}
	if len(result.Records) != len(expectedTargets) {
		t.Fatalf("Expected %d records, got %d", len(expectedTargets), len(result.Records))
	}

	for i, expectedTarget := range expectedTargets {
		if result.Records[i] != expectedTarget {
			t.Errorf("Expected CNAME %s at index %d, got %s", expectedTarget, i, result.Records[i])
		}
	}
}

func TestClient_Query_NoCNAMERecords(t *testing.T) {
	// Create mock DNS server that returns successful response but no CNAME records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		// No answer records added

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("alias.example.com", "CNAME", serverAddr)

	if err == nil {
		t.Fatal("Expected error for no CNAME records")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "no CNAME records found") {
		t.Errorf("Expected error about no CNAME records, got: %s", err.Error())
	}
}

// Tests for TXT record support

func TestClient_Query_SuccessfulTXTRecord(t *testing.T) {
	// Create mock DNS server that returns TXT record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Add small delay to ensure query time is measurable
		time.Sleep(1 * time.Millisecond)

		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add TXT record to response
		txtRecord := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Txt: []string{"v=spf1 include:_spf.example.com ~all"},
		}
		msg.Answer = append(msg.Answer, txtRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "TXT", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %s", result.Domain)
	}

	if result.RecordType != "TXT" {
		t.Errorf("Expected record type 'TXT', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "v=spf1 include:_spf.example.com ~all" {
		t.Errorf("Expected TXT 'v=spf1 include:_spf.example.com ~all', got %s", result.Records[0])
	}

	if result.QueryTime <= 0 {
		t.Error("Expected positive query time")
	}

	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_TXTRecordWithMultipleStrings(t *testing.T) {
	// Create mock DNS server that returns TXT record with multiple strings
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add TXT record with multiple strings to response
		txtRecord := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			Txt: []string{"part1", "part2", "part3"},
		}
		msg.Answer = append(msg.Answer, txtRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "TXT", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	// Multiple strings should be joined with spaces
	if result.Records[0] != "part1 part2 part3" {
		t.Errorf("Expected TXT 'part1 part2 part3', got %s", result.Records[0])
	}
}

func TestClient_Query_MultipleTXTRecords(t *testing.T) {
	// Create mock DNS server that returns multiple TXT records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add multiple TXT records to response
		txtRecords := [][]string{
			{"v=spf1 include:_spf.example.com ~all"},
			{"google-site-verification=abcd1234"},
			{"keybase-site-verification=xyz789"},
		}

		for _, txt := range txtRecords {
			txtRecord := &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    300,
				},
				Txt: txt,
			}
			msg.Answer = append(msg.Answer, txtRecord)
		}

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "TXT", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	expectedTXTs := []string{
		"v=spf1 include:_spf.example.com ~all",
		"google-site-verification=abcd1234",
		"keybase-site-verification=xyz789",
	}
	if len(result.Records) != len(expectedTXTs) {
		t.Fatalf("Expected %d records, got %d", len(expectedTXTs), len(result.Records))
	}

	for i, expectedTXT := range expectedTXTs {
		if result.Records[i] != expectedTXT {
			t.Errorf("Expected TXT %s at index %d, got %s", expectedTXT, i, result.Records[i])
		}
	}
}

func TestClient_Query_NoTXTRecords(t *testing.T) {
	// Create mock DNS server that returns successful response but no TXT records
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true
		// No answer records added

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "TXT", serverAddr)

	if err == nil {
		t.Fatal("Expected error for no TXT records")
	}

	if !errors.IsDNSError(err) {
		t.Errorf("Expected DNS error, got %T", err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	if !strings.Contains(err.Error(), "no TXT records found") {
		t.Errorf("Expected error about no TXT records, got: %s", err.Error())
	}
}

// Tests for custom DNS server functionality

func TestClient_Query_CustomDNSServer(t *testing.T) {
	// Create mock DNS server
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add A record to response
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("192.0.2.100"),
		}
		msg.Answer = append(msg.Answer, aRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	// Verify the custom server was used
	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "192.0.2.100" {
		t.Errorf("Expected IP '192.0.2.100', got %s", result.Records[0])
	}
}

func TestClient_Query_CustomDNSServerWithPort(t *testing.T) {
	// Create mock DNS server
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add A record to response
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("192.0.2.101"),
		}
		msg.Answer = append(msg.Answer, aRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()
	result, err := client.Query("example.com", "A", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	// Verify the custom server with port was used
	if result.Server != serverAddr {
		t.Errorf("Expected server %s, got %s", serverAddr, result.Server)
	}
}

func TestClient_Query_CustomDNSServerWithoutPort(t *testing.T) {
	// For this test, we'll use a known public DNS server without port
	// to test that the default port is added correctly
	client := NewClient()

	// Use Google DNS without port
	result, _ := client.Query("example.com", "A", "8.8.8.8")

	// Verify the server with default port was used
	expectedServer := "8.8.8.8:53"
	if result.Server != expectedServer {
		t.Errorf("Expected server %s, got %s", expectedServer, result.Server)
	}
}

func TestClient_Query_InvalidCustomDNSServer(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name   string
		server string
	}{
		{"invalid IP format", "invalid.server"},
		{"empty server", ""},
		{"invalid port format", "127.0.0.1:invalid"},
		{"non-IP hostname", "not-an-ip.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty server test as it should use system default
			if tt.server == "" {
				return
			}

			result, err := client.Query("example.com", "A", tt.server)

			if err == nil {
				t.Fatal("Expected error for invalid DNS server")
			}

			if !errors.IsInputError(err) {
				t.Errorf("Expected input error, got %T: %v", err, err)
			}

			if result.Error == nil {
				t.Fatal("Expected result.Error to be set")
			}
		})
	}
}

func TestClient_Query_UnreachableDNSServer(t *testing.T) {
	client := NewClient()
	client.SetTimeout(100 * time.Millisecond) // Short timeout

	// Use a non-routable IP address (RFC 5737 test address)
	unreachableServer := "192.0.2.254:53"

	result, err := client.Query("example.com", "A", unreachableServer)

	if err == nil {
		t.Fatal("Expected error for unreachable DNS server")
	}

	if !errors.IsNetworkError(err) {
		t.Errorf("Expected network error, got %T: %v", err, err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	// Should be classified as a timeout or connection error
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "timeout") && !strings.Contains(errStr, "unreachable") && !strings.Contains(errStr, "refused") {
		t.Errorf("Expected timeout/unreachable/refused error, got: %s", err.Error())
	}
}

func TestClient_Query_SystemDefaultDNS(t *testing.T) {
	client := NewClient()

	// Query with empty server should use system default
	result, err := client.Query("example.com", "A", "")

	// We don't expect this to necessarily succeed since we're using system DNS
	// but we want to verify that a server was selected
	if result.Server == "" {
		t.Error("Expected a DNS server to be selected when using system default")
	}

	// The server should either be the system default or the fallback (8.8.8.8:53)
	if result.Server != "8.8.8.8:53" {
		// If it's not the fallback, it should be a valid IP:port format
		host, port, splitErr := net.SplitHostPort(result.Server)
		if splitErr != nil {
			t.Errorf("Server should be in IP:port format, got: %s", result.Server)
		} else {
			if net.ParseIP(host) == nil {
				t.Errorf("Server host should be a valid IP, got: %s", host)
			}
			if port != "53" {
				t.Errorf("Expected port 53, got: %s", port)
			}
		}
	}

	// Log the result for debugging (this test might fail in some environments)
	t.Logf("System DNS result: server=%s, error=%v", result.Server, err)
}

func TestClient_Query_DNSServerConnectionRefused(t *testing.T) {
	// Create a TCP listener that immediately closes connections to simulate connection refused
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Get the address and convert to UDP format for DNS
	tcpAddr := listener.Addr().String()
	host, port, err := net.SplitHostPort(tcpAddr)
	if err != nil {
		t.Fatalf("Failed to split address: %v", err)
	}

	// Close the listener to ensure connection refused
	listener.Close()

	client := NewClient()
	client.SetTimeout(1 * time.Second)

	// Try to query the closed port (this should result in connection refused)
	result, err := client.Query("example.com", "A", host+":"+port)

	if err == nil {
		t.Fatal("Expected error for connection refused")
	}

	if !errors.IsNetworkError(err) {
		t.Errorf("Expected network error, got %T: %v", err, err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}
}

func TestClient_Query_DNSServerTimeout(t *testing.T) {
	// Create mock DNS server that doesn't respond (simulates timeout)
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		// Don't respond to simulate timeout
		time.Sleep(2 * time.Second)
	})
	defer cleanup()

	client := NewClient()
	client.SetTimeout(100 * time.Millisecond) // Very short timeout

	start := time.Now()
	result, err := client.Query("example.com", "A", serverAddr)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error")
	}

	if !errors.IsNetworkError(err) {
		t.Errorf("Expected network error, got %T: %v", err, err)
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}

	// Should timeout quickly
	if elapsed > 500*time.Millisecond {
		t.Errorf("Query took too long: %v", elapsed)
	}

	// Should be classified as a timeout error
	if !strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Errorf("Expected timeout error, got: %s", err.Error())
	}
}

func TestClient_Query_ValidDNSServerFormats(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name           string
		server         string
		expectedServer string
	}{
		{"IPv4 with port", "8.8.8.8:53", "8.8.8.8:53"},
		{"IPv4 without port", "8.8.8.8", "8.8.8.8:53"},
		{"IPv6 with port", "[2001:4860:4860::8888]:53", "[2001:4860:4860::8888]:53"},
		{"localhost IPv4", "127.0.0.1", "127.0.0.1:53"},
		{"localhost IPv6", "::1", "[::1]:53"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := client.Query("example.com", "A", tt.server)

			// We don't care if the query succeeds or fails, just that the server format is correct
			if result.Server != tt.expectedServer {
				t.Errorf("Expected server %s, got %s", tt.expectedServer, result.Server)
			}
		})
	}
}

func TestClient_Query_DNSServerValidation(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name   string
		server string
		valid  bool
	}{
		{"valid IPv4", "8.8.8.8", true},
		{"valid IPv4 with port", "8.8.8.8:53", true},
		{"valid IPv6", "2001:4860:4860::8888", true},
		{"valid IPv6 with port", "[2001:4860:4860::8888]:53", true},
		{"invalid IP", "999.999.999.999", false},
		{"invalid format", "not-an-ip", false},
		{"invalid port", "8.8.8.8:99999", false},
		{"empty string", "", true}, // Empty should use system default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Query("example.com", "A", tt.server)

			if tt.valid {
				// For valid servers, we should not get input validation errors
				if err != nil && errors.IsInputError(err) {
					t.Errorf("Expected valid server %s to pass validation, got: %v", tt.server, err)
				}
			} else {
				// For invalid servers, we should get input validation errors
				if err == nil || !errors.IsInputError(err) {
					t.Errorf("Expected invalid server %s to fail validation, got error: %v", tt.server, err)
				}
				if result.Error == nil {
					t.Error("Expected result.Error to be set for invalid server")
				}
			}
		})
	}
}

// Tests for case insensitive record types

func TestClient_Query_CaseInsensitiveRecordTypes(t *testing.T) {
	// Create mock DNS server that returns A record
	serverAddr, cleanup := mockDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Authoritative = true

		// Add A record to response
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("192.0.2.1"),
		}
		msg.Answer = append(msg.Answer, aRecord)

		w.WriteMsg(msg)
	})
	defer cleanup()

	client := NewClient()

	// Test lowercase record type
	result, err := client.Query("example.com", "a", serverAddr)

	if err != nil {
		t.Fatalf("Unexpected error for lowercase 'a': %v", err)
	}

	if result.Error != nil {
		t.Fatalf("Expected no error in result, got: %v", result.Error)
	}

	if result.RecordType != "a" {
		t.Errorf("Expected record type 'a', got %s", result.RecordType)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0] != "192.0.2.1" {
		t.Errorf("Expected IP '192.0.2.1', got %s", result.Records[0])
	}
}

// Test mixed case scenarios for all record types
func TestClient_Query_AllRecordTypesCaseInsensitive(t *testing.T) {
	testCases := []struct {
		recordType string
		expected   string
	}{
		{"a", "A"},
		{"A", "A"},
		{"aaaa", "AAAA"},
		{"AAAA", "AAAA"},
		{"mx", "MX"},
		{"MX", "MX"},
		{"cname", "CNAME"},
		{"CNAME", "CNAME"},
		{"txt", "TXT"},
		{"TXT", "TXT"},
	}

	for _, tc := range testCases {
		t.Run("record_type_"+tc.recordType, func(t *testing.T) {
			client := NewClient()

			// We expect this to fail with network error since we're not providing a mock server
			// But it should not fail with input validation error
			result, err := client.Query("example.com", tc.recordType, "8.8.8.8")

			// The record type should be preserved as provided by user
			if result.RecordType != tc.recordType {
				t.Errorf("Expected record type '%s', got %s", tc.recordType, result.RecordType)
			}

			// Should not be an input validation error
			if err != nil && errors.IsInputError(err) && strings.Contains(err.Error(), "unsupported record type") {
				t.Errorf("Record type '%s' should be supported, got error: %s", tc.recordType, err.Error())
			}
		})
	}
}
