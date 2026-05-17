package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// This test builds the binary and runs it with different flags
// to ensure they are all recognized and have a basic effect.
func TestMainFlags(t *testing.T) {
	// Build the binary to a temporary file for testing
	binaryName := "go-xbuild-go-test"
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary for testing: %v", err)
	}
	// Clean up the binary after the tests are done
	defer os.Remove(binaryName)

	testCases := []struct {
		name           string   // Name of the test
		args           []string // Arguments to pass to the binary
		expectedOutput string   // A substring we expect in the output
		expectFail     bool     // Whether we expect the command to exit with an error
	}{
		{
			name:           "--help flag",
			args:           []string{"--help"},
			expectedOutput: "Usage:",
			expectFail:     false, // Exits with status 0, but exec.Command sees that as a non-error exit. We expect it not to fail. Let's adjust.
		},
		{
			name:           "--version flag",
			args:           []string{"--version"},
			expectedOutput: "go-xbuild-go", // Should print the program name and version
		},
		{
			name:           "invalid flag",
			args:           []string{"--this-flag-does-not-exist"},
			expectedOutput: "flag provided but not defined",
			expectFail:     true,
		},
		{
			name:           "--pi flag exists",
			args:           []string{"--pi=false", "--help"}, // Check if the flag is recognized
			expectedOutput: "Usage:",
		},
		{
			name:           "--platforms-file flag exists",
			args:           []string{"--platforms-file=test.txt", "--help"}, // Check if the flag is recognized
			expectedOutput: "Usage:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./"+binaryName, tc.args...)
			output, err := cmd.CombinedOutput()

			if tc.expectFail && err == nil {
				t.Errorf("Expected command to fail, but it succeeded.")
			}
			if !tc.expectFail && err != nil {
				t.Errorf("Expected command to succeed, but it failed with: %v", err)
			}
			
			if !strings.Contains(string(output), tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', but got:\n%s", tc.expectedOutput, string(output))
			}
		})
	}
}
