package build

import (
	"reflect"
	"testing"
)

func TestParseArguments(t *testing.T) {
	// Define our test cases
	testCases := []struct {
		name          string // The name of the test case
		input         string // The string we'll pass to the function
		expected      []string // The slice we expect to get back
		expectError   bool   // Whether we expect an error
	}{
		{
			name:          "Simple arguments",
			input:         "-tags systray -race",
			expected:      []string{"-tags", "systray", "-race"},
			expectError:   false,
		},
		{
			name:          "Arguments with quotes",
			input:         `-ldflags "-X main.version=1.2.3 -s -w"`,
			expected:      []string{"-ldflags", "-X main.version=1.2.3 -s -w"},
			expectError:   false,
		},
		{
			name:          "Empty input string",
			input:         "",
			expected:      nil, // Expect nil, not an empty slice
			expectError:   false,
		},
		{
			name:          "Input with only spaces",
			input:         "   ",
			expected:      nil,
			expectError:   false,
		},
		{
			name:          "Unclosed quote",
			input:         `-ldflags "-X main.version=1.2.3`,
			expected:      nil,
			expectError:   true,
		},
	}

	// Loop over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the function we're testing
			actual, err := ParseArguments(tc.input)

			// Check if an error occurred when we didn't expect one
			if err != nil && !tc.expectError {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			// Check if an error was expected but didn't occur
			if err == nil && tc.expectError {
				t.Fatalf("Expected an error, but got none")
			}
			
			// Check if the actual output matches the expected output
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}