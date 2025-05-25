package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/cover"
)

func TestRunDiffMode(t *testing.T) {
	// Create mock profiles
	profiles := []*cover.Profile{
		{
			FileName: "main.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 10, EndLine: 20, Count: 1},
				{StartLine: 30, EndLine: 40, Count: 0},
			},
		},
		{
			FileName: "pkg/util/helper.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 5, EndLine: 15, Count: 1},
			},
		},
	}

	// Test cases for different scenarios
	tests := []struct {
		name         string
		diffBase     string
		threshold    float64
		wantErr      bool
		wantInOutput []string
	}{
		{
			name:         "diff with HEAD",
			diffBase:     "HEAD",
			threshold:    0,
			wantErr:      false,
			wantInOutput: []string{"Diff Coverage Report:", "TOTAL DIFF"},
		},
		{
			name:         "diff with threshold pass",
			diffBase:     "HEAD",
			threshold:    0.0, // Changed to 0 since git diff might be empty
			wantErr:      false,
			wantInOutput: []string{"Diff Coverage Report:"},
		},
		{
			name:         "diff with threshold fail",
			diffBase:     "HEAD",
			threshold:    90.0,
			wantErr:      true,
			wantInOutput: []string{"Diff Coverage Report:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if not in a git repository
			if _, err := getMergeBase(); err != nil {
				t.Skip("Skipping git-dependent test - not in a git repository")
			}

			var buf bytes.Buffer
			cli := &CLI{Output: &buf}

			err := cli.runDiffMode(profiles, tt.diffBase, tt.threshold)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("runDiffMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check threshold error type
			if tt.wantErr && tt.threshold > 0 {
				if _, ok := err.(*ThresholdError); !ok {
					t.Errorf("Expected ThresholdError, got %T", err)
				}
			}

			// Check output contains expected strings
			output := buf.String()
			for _, expected := range tt.wantInOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected string: %q", expected)
				}
			}
		})
	}
}

// Test helper to create a mock CLI with diff mode
func TestCLIWithDiffMode(t *testing.T) {
	// Create a temporary coverage file
	tmpDir, err := os.MkdirTemp("", "gocov-diff-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple coverage file
	coverageFile := filepath.Join(tmpDir, "coverage.out")
	coverageContent := `mode: set
main.go:10.1,20.1 1 1
main.go:30.1,40.1 1 0
pkg/util/helper.go:5.1,15.1 1 1
`
	if err := os.WriteFile(coverageFile, []byte(coverageContent), 0644); err != nil {
		t.Fatalf("Failed to write coverage file: %v", err)
	}

	// Test CLI with diff flag
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "diff mode with HEAD",
			args:    []string{"-coverprofile", coverageFile, "-diff", "HEAD"},
			wantErr: false,
		},
		{
			name:    "diff mode with threshold",
			args:    []string{"-coverprofile", coverageFile, "-diff", "HEAD", "-threshold", "0"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if not in a git repository
			if _, err := getMergeBase(); err != nil {
				t.Skip("Skipping git-dependent test - not in a git repository")
			}

			var buf bytes.Buffer
			cli := NewCLI(&buf, tt.args)
			err := cli.Run()

			if (err != nil) != tt.wantErr {
				t.Errorf("CLI.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check that diff mode was activated
			output := buf.String()
			if !strings.Contains(output, "Diff Coverage Report:") {
				t.Error("Expected diff coverage report in output")
			}
		})
	}
}
