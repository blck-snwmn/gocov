package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestThresholdCheck(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "gocov-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test coverage file with known coverage
	coverageFile := filepath.Join(tmpDir, "coverage.out")
	coverageContent := `mode: set
github.com/example/project/main.go:1.1,2.1 1 1
github.com/example/project/main.go:3.1,4.1 1 1
github.com/example/project/main.go:5.1,6.1 1 0
github.com/example/project/main.go:7.1,8.1 1 0
github.com/example/project/main.go:9.1,10.1 1 0
`
	if err := os.WriteFile(coverageFile, []byte(coverageContent), 0644); err != nil {
		t.Fatalf("Failed to write coverage file: %v", err)
	}

	// Test cases
	tests := []struct {
		name      string
		threshold float64
		wantErr   bool
		errType   string
	}{
		{
			name:      "threshold passed",
			threshold: 30.0, // Coverage is 40%, so this should pass
			wantErr:   false,
		},
		{
			name:      "threshold failed",
			threshold: 50.0, // Coverage is 40%, so this should fail
			wantErr:   true,
			errType:   "ThresholdError",
		},
		{
			name:      "no threshold",
			threshold: 0.0, // No threshold check
			wantErr:   false,
		},
		{
			name:      "exact threshold",
			threshold: 40.0, // Exact match should pass
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			args := []string{
				"-coverprofile", coverageFile,
			}
			if tc.threshold > 0 {
				args = append(args, "-threshold", formatFloat(tc.threshold))
			}

			cli := NewCLI(&buf, args)
			err := cli.Run()

			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errType == "ThresholdError" {
					if _, ok := err.(*ThresholdError); !ok {
						t.Errorf("Expected ThresholdError but got: %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		wantErr   bool
	}{
		{
			name:      "valid threshold 0",
			threshold: 0.0,
			wantErr:   false,
		},
		{
			name:      "valid threshold 50",
			threshold: 50.0,
			wantErr:   false,
		},
		{
			name:      "valid threshold 100",
			threshold: 100.0,
			wantErr:   false,
		},
		{
			name:      "invalid threshold negative",
			threshold: -1.0,
			wantErr:   true,
		},
		{
			name:      "invalid threshold over 100",
			threshold: 101.0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateThreshold(tc.threshold)
			if tc.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestThresholdError(t *testing.T) {
	err := NewThresholdError(80.0, 75.5)
	thresholdErr, ok := err.(*ThresholdError)
	if !ok {
		t.Fatalf("Expected ThresholdError but got %T", err)
	}

	expectedMsg := "coverage 75.5% is below threshold 80.0%"
	if thresholdErr.Error() != expectedMsg {
		t.Errorf("Expected error message %q but got %q", expectedMsg, thresholdErr.Error())
	}
}

// Helper function to format float as string
func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
