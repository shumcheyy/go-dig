package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-dig/cmd"
	"go-dig/pkg/dns"
	"go-dig/pkg/errors"
	"go-dig/pkg/output"
)

func main() {
	// Create components
	parser := cmd.NewCLIParser()
	formatter := output.NewFormatter()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle system-level panics and errors
	defer func() {
		if r := recover(); r != nil {
			systemErr := errors.NewSystemError(fmt.Sprintf("application panic: %v", r), nil)
			fmt.Fprint(os.Stderr, formatter.FormatError(systemErr))
			os.Exit(getExitCode(systemErr))
		}
	}()

	// Handle signals in a separate goroutine
	go func() {
		sig := <-sigChan
		systemErr := errors.NewSystemError(fmt.Sprintf("received signal %v, shutting down gracefully", sig), nil)
		fmt.Fprint(os.Stderr, formatter.FormatError(systemErr))
		os.Exit(130) // Standard exit code for SIGINT
	}()

	// Parse command-line arguments
	config, err := parser.Parse(os.Args[1:])
	if err != nil {
		// Use the enhanced error formatter
		fmt.Fprint(os.Stderr, formatter.FormatError(err))

		// Show usage for input errors
		if errors.IsInputError(err) {
			fmt.Fprintf(os.Stderr, "\n")
			parser.ShowUsage()
		}

		// Exit with appropriate code based on error type
		os.Exit(getExitCode(err))
	}

	// Create DNS client
	client := dns.NewClient()
	client.SetTimeout(config.Timeout)

	// Perform DNS query with proper error propagation
	result, err := client.Query(config.Domain, config.RecordType, config.Server)
	if err != nil {
		// Ensure error is properly formatted and propagated
		fmt.Fprint(os.Stderr, formatter.FormatError(err))
		os.Exit(getExitCode(err))
	}

	// Validate result before formatting
	if result == nil {
		systemErr := errors.NewSystemError("DNS query returned nil result", nil)
		fmt.Fprint(os.Stderr, formatter.FormatError(systemErr))
		os.Exit(getExitCode(systemErr))
	}

	// Format and display results
	output := formatter.FormatResult(result)
	fmt.Print(output)

	// Successful execution
	os.Exit(0)
}

// getExitCode returns appropriate exit code based on error type
// Exit codes follow standard conventions:
// 0 = Success
// 1 = General error / Invalid arguments
// 2 = Network/DNS error
// 3 = System error
// 130 = Interrupted by signal (SIGINT)
func getExitCode(err error) int {
	if err == nil {
		return 0
	}

	if digErr, ok := err.(*errors.DigError); ok {
		switch digErr.Type {
		case errors.ErrorTypeInput:
			return 1 // Invalid arguments
		case errors.ErrorTypeNetwork:
			return 2 // Network error
		case errors.ErrorTypeDNS:
			return 2 // DNS error
		case errors.ErrorTypeSystem:
			return 3 // System error
		}
	}
	return 1 // Default error code for unknown errors
}
