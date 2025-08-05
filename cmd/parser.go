package cmd

import (
"flag"
"fmt"
"os"
"strings"
"time"

"go-dig/pkg/errors"
)

// Config holds the parsed command-line configuration
type Config struct {
Domain     string
RecordType string
Server     string
Timeout    time.Duration
}

// Parser interface defines the contract for CLI argument parsing
type Parser interface {
Parse(args []string) (*Config, error)
ShowUsage()
}

// CLIParser implements the Parser interface
type CLIParser struct{}

// NewCLIParser creates a new CLI parser instance
func NewCLIParser() Parser {
return &CLIParser{}
}

// Parse parses command-line arguments and returns a Config struct
func (p *CLIParser) Parse(args []string) (*Config, error) {
config := &Config{
RecordType: "A",             // Default record type
Timeout:    5 * time.Second, // Default timeout
}

// Create a new flag set for each parse operation to avoid conflicts
flagSet := flag.NewFlagSet("go-dig", flag.ContinueOnError)

// Define flags
recordType := flagSet.String("t", "A", "DNS record type (A, AAAA, MX, CNAME, TXT)")
server := flagSet.String("s", "", "DNS server to use (IP address)")

// Suppress default error output from flag package
flagSet.SetOutput(os.Stderr)

// Parse flags
err := flagSet.Parse(args)
if err != nil {
return nil, errors.NewInputError("invalid command line arguments", err)
}

// Get remaining arguments (should be the domain)
remaining := flagSet.Args()
if len(remaining) == 0 {
return nil, errors.NewInputError("domain name is required", nil)
}
if len(remaining) > 1 {
return nil, errors.NewInputError(fmt.Sprintf("too many arguments: expected domain name only, got %d arguments", len(remaining)), nil)
}

config.Domain = remaining[0]
config.RecordType = strings.ToUpper(*recordType)
config.Server = *server

// Check if -s flag was explicitly provided
serverFlagProvided := false
flagSet.Visit(func(f *flag.Flag) {
if f.Name == "s" {
serverFlagProvided = true
}
})

// Validate inputs using the new error handling
if err := p.validateConfig(config, serverFlagProvided); err != nil {
return nil, err
}

return config, nil
}

// validateConfig validates the parsed configuration
func (p *CLIParser) validateConfig(config *Config, serverFlagProvided bool) error {
// Validate domain name using the new error handling
if err := errors.ValidateDomain(config.Domain); err != nil {
return err
}

// Validate record type using the new error handling
if err := errors.ValidateRecordType(config.RecordType); err != nil {
return err
}

// Validate DNS server if flag was provided
if serverFlagProvided {
if err := errors.ValidateDNSServer(config.Server); err != nil {
return err
}
}

return nil
}

// ShowUsage displays usage information
func (p *CLIParser) ShowUsage() {
fmt.Fprintf(os.Stderr, "Usage: go-dig <domain> [options]\n\n")
fmt.Fprintf(os.Stderr, "Arguments:\n")
fmt.Fprintf(os.Stderr, "  domain       Domain name to query\n\n")
fmt.Fprintf(os.Stderr, "Options:\n")
fmt.Fprintf(os.Stderr, "  -t <type>    DNS record type (A, AAAA, MX, CNAME, TXT) [default: A]\n")
fmt.Fprintf(os.Stderr, "  -s <server>  DNS server to use (IP address) [default: system default]\n\n")
fmt.Fprintf(os.Stderr, "Examples:\n")
fmt.Fprintf(os.Stderr, "  go-dig google.com\n")
fmt.Fprintf(os.Stderr, "  go-dig google.com -t AAAA\n")
fmt.Fprintf(os.Stderr, "  go-dig google.com -s 8.8.8.8\n")
fmt.Fprintf(os.Stderr, "  go-dig google.com -t MX -s 1.1.1.1\n")
}
