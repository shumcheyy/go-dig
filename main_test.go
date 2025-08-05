package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainApplicationFlow(t *testing.T) {
	// Build the application first
	buildCmd := exec.Command("go", "build", "-o", "go-dig-test.exe")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("go-dig-test.exe")

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedExit   int
		containsOutput []string
	}{
		{
			name:           "Basic A record query",
			args:           []string{"google.com"},
			expectError:    false,
			expectedExit:   0,
			containsOutput: []string{"google.com", "IN", "A", "ANSWER SECTION"},
		},
		{
			name:           "AAAA record query",
			args:           []string{"-t", "AAAA", "google.com"},
			expectError:    false,
			expectedExit:   0,
			containsOutput: []string{"google.com", "IN", "AAAA", "ANSWER SECTION"},
		},
		{
			name:           "MX record query",
			args:           []string{"-t", "MX", "google.com"},
			expectError:    false,
			expectedExit:   0,
			containsOutput: []string{"google.com", "IN", "MX", "ANSWER SECTION"},
		},
		{
			name:           "Custom DNS server",
			args:           []string{"-s", "8.8.8.8", "google.com"},
			expectError:    false,
			expectedExit:   0,
			containsOutput: []string{"google.com", "8.8.8.8:53", "ANSWER SECTION"},
		},
		{
			name:           "No arguments",
			args:           []string{},
			expectError:    true,
			expectedExit:   1,
			containsOutput: []string{"domain name is required", "Usage:"},
		},
		{
			name:           "Invalid record type",
			args:           []string{"-t", "INVALID", "google.com"},
			expectError:    true,
			expectedExit:   1,
			containsOutput: []string{"unsupported record type", "INVALID"},
		},
		{
			name:           "Invalid DNS server",
			args:           []string{"-s", "invalid.server", "google.com"},
			expectError:    true,
			expectedExit:   1,
			containsOutput: []string{"not a valid IP address"},
		},
		{
			name:           "Non-existent domain",
			args:           []string{"nonexistentdomain12345.com"},
			expectError:    true,
			expectedExit:   2,
			containsOutput: []string{"not found", "NXDOMAIN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./go-dig-test.exe", tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode := exitErr.ExitCode()
				if exitCode != tt.expectedExit {
					t.Errorf("Expected exit code %d, got %d", tt.expectedExit, exitCode)
				}
			} else if err != nil && tt.expectError {
				t.Errorf("Expected command to fail but got unexpected error: %v", err)
			} else if err == nil && tt.expectError {
				t.Errorf("Expected command to fail but it succeeded")
			} else if err != nil && !tt.expectError {
				t.Errorf("Expected command to succeed but it failed: %v", err)
			}

			// Check output contains expected strings
			for _, expected := range tt.containsOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Output does not contain expected string '%s'\nOutput: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "No error",
			err:      nil,
			expected: 0,
		},
		{
			name:     "Non-DigError",
			err:      &testError{},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExitCode(tt.err)
			if result != tt.expected {
				t.Errorf("Expected exit code %d, got %d", tt.expected, result)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}
