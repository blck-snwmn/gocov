package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMainE2E tests the main function end-to-end
func TestMainE2E(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "gocov_test", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("gocov_test")

	tests := []struct {
		name        string
		args        []string
		wantError   bool
		wantOutput  []string
		wantMissing []string
	}{
		{
			name:      "no arguments",
			args:      []string{},
			wantError: true,
		},
		{
			name:       "help flag",
			args:       []string{"-h"},
			wantError:  true,
			wantOutput: []string{"coverprofile"},
		},
		{
			name:       "valid coverage file",
			args:       []string{"-coverprofile", "testdata/coverage.out"},
			wantError:  false,
			wantOutput: []string{"Directory", "TOTAL", "github.com/example/project"},
		},
		{
			name:       "with level flag",
			args:       []string{"-coverprofile", "testdata/coverage.out", "-level", "3"},
			wantError:  false,
			wantOutput: []string{"github.com/example/project", "TOTAL"},
		},
		{
			name:       "with min coverage filter",
			args:       []string{"-coverprofile", "testdata/coverage.out", "-min", "80"},
			wantError:  false,
			wantOutput: []string{"internal/service", "FILTERED TOTAL"},
		},
		{
			name:       "with json format",
			args:       []string{"-coverprofile", "testdata/coverage.out", "-format", "json"},
			wantError:  false,
			wantOutput: []string{`"results"`, `"total"`, `"coverage"`},
		},
		{
			name:      "invalid coverage file",
			args:      []string{"-coverprofile", "testdata/nonexistent.out"},
			wantError: true,
		},
		{
			name:      "invalid format",
			args:      []string{"-coverprofile", "testdata/coverage.out", "-format", "xml"},
			wantError: true,
		},
		{
			name:      "invalid min coverage",
			args:      []string{"-coverprofile", "testdata/coverage.out", "-min", "150"},
			wantError: true,
		},
		{
			name:      "min greater than max",
			args:      []string{"-coverprofile", "testdata/coverage.out", "-min", "80", "-max", "50"},
			wantError: true,
		},
		{
			name:        "with ignore patterns",
			args:        []string{"-coverprofile", "testdata/coverage.out", "-ignore", "*/internal/*"},
			wantError:   false,
			wantOutput:  []string{"pkg/util", "cmd/server"},
			wantMissing: []string{"internal/service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./gocov_test", tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			hasError := err != nil

			if hasError != tt.wantError {
				t.Errorf("error = %v, wantError = %v\nstderr: %s", err, tt.wantError, stderr.String())
			}

			output := stdout.String() + stderr.String()

			// Check expected output
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q\nGot: %s", want, output)
				}
			}

			// Check missing output
			for _, missing := range tt.wantMissing {
				if strings.Contains(output, missing) {
					t.Errorf("output should not contain %q\nGot: %s", missing, output)
				}
			}
		})
	}
}

// TestMainWithConfig tests main function with configuration file
func TestMainWithConfig(t *testing.T) {
	// Create a temporary directory and config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".gocov.yml")

	configContent := `level: 3
coverage:
  min: 50
  max: 90
format: json
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Change to temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Build the binary in the original directory
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "gocov_test_config"), ".")
	buildCmd.Dir = originalWd
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(filepath.Join(tempDir, "gocov_test_config"))

	// Run the binary with config
	cmd := exec.Command("./gocov_test_config", "-coverprofile", filepath.Join(originalWd, "testdata/coverage.out"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run with config: %v\nstderr: %s", err, stderr.String())
	}

	output := stdout.String()
	// Should use JSON format from config
	if !strings.Contains(output, `"results"`) {
		t.Errorf("Expected JSON output from config file\nGot: %s", output)
	}
	// Should use level 3 from config
	if !strings.Contains(output, "github.com/example/project") {
		t.Errorf("Expected aggregated directory from config level\nGot: %s", output)
	}
}
