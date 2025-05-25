package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != 0 {
		t.Errorf("Expected default level to be 0, got %d", config.Level)
	}
	if config.Coverage.Min != 0.0 {
		t.Errorf("Expected default min coverage to be 0.0, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 100.0 {
		t.Errorf("Expected default max coverage to be 100.0, got %f", config.Coverage.Max)
	}
	if config.Format != "table" {
		t.Errorf("Expected default format to be 'table', got %s", config.Format)
	}
	if len(config.Ignore) != 0 {
		t.Errorf("Expected default ignore to be empty, got %v", config.Ignore)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".gocov.yml")

	configContent := `
level: 2
coverage:
  min: 50
  max: 90
format: json
ignore:
  - "*/vendor/*"
  - "*/test/*"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading config
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be loaded, got nil")
	}

	if config.Level != 2 {
		t.Errorf("Expected level to be 2, got %d", config.Level)
	}
	if config.Coverage.Min != 50 {
		t.Errorf("Expected min coverage to be 50, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 90 {
		t.Errorf("Expected max coverage to be 90, got %f", config.Coverage.Max)
	}
	if config.Format != "json" {
		t.Errorf("Expected format to be 'json', got %s", config.Format)
	}
	if len(config.Ignore) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(config.Ignore))
	}
}

func TestLoadConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		wantErrType error
	}{
		{
			name: "invalid min coverage below 0",
			configYAML: `coverage:
  min: -10
  max: 100
`,
			wantErrType: &ValidationError{},
		},
		{
			name: "invalid min coverage above 100",
			configYAML: `coverage:
  min: 150
  max: 100
`,
			wantErrType: &ValidationError{},
		},
		{
			name: "invalid max coverage below 0",
			configYAML: `coverage:
  min: 0
  max: -10
`,
			wantErrType: &ValidationError{},
		},
		{
			name: "invalid max coverage above 100",
			configYAML: `coverage:
  min: 0
  max: 150
`,
			wantErrType: &ValidationError{},
		},
		{
			name: "min greater than max",
			configYAML: `coverage:
  min: 80
  max: 50
`,
			wantErrType: &ValidationError{},
		},
		{
			name:        "invalid format",
			configYAML:  `format: xml`,
			wantErrType: &ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, ".gocov.yml")

			if err := os.WriteFile(configFile, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			_, err := LoadConfig(configFile)
			if err == nil {
				t.Error("Expected error, got nil")
			} else if tt.wantErrType != nil {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	config, err := LoadConfig("non-existent-file.yml")
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}
	if config != nil {
		t.Error("Expected nil config for non-existent file")
	}
}

func TestMergeWithFlags(t *testing.T) {
	config := DefaultConfig()

	// Set up flag values
	level := 3
	minCoverage := 60.0
	maxCoverage := 95.0
	outputFormat := "json"
	ignorePatterns := []string{"*/vendor/*"}

	// Test merging when flags have non-default values
	config.MergeWithFlags(&level, &minCoverage, &maxCoverage, &outputFormat, ignorePatterns)

	if config.Level != 3 {
		t.Errorf("Expected level to be 3 after merge, got %d", config.Level)
	}
	if config.Coverage.Min != 60.0 {
		t.Errorf("Expected min coverage to be 60.0 after merge, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 95.0 {
		t.Errorf("Expected max coverage to be 95.0 after merge, got %f", config.Coverage.Max)
	}
	if config.Format != "json" {
		t.Errorf("Expected format to be 'json' after merge, got %s", config.Format)
	}
	if len(config.Ignore) != 1 || config.Ignore[0] != "*/vendor/*" {
		t.Errorf("Expected ignore patterns to be updated after merge, got %v", config.Ignore)
	}

	// Reset config
	config = &Config{
		Level: 5,
		Coverage: CoverageConfig{
			Min: 70.0,
			Max: 80.0,
		},
		Format: "table",
		Ignore: []string{"*/test/*"},
	}

	// Test merging when flags have default values (should not override config)
	level = 0
	minCoverage = 0.0
	maxCoverage = 100.0
	outputFormat = "table"
	ignorePatterns = nil

	config.MergeWithFlags(&level, &minCoverage, &maxCoverage, &outputFormat, ignorePatterns)

	if config.Level != 5 {
		t.Errorf("Expected level to remain 5, got %d", config.Level)
	}
	if config.Coverage.Min != 70.0 {
		t.Errorf("Expected min coverage to remain 70.0, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 80.0 {
		t.Errorf("Expected max coverage to remain 80.0, got %f", config.Coverage.Max)
	}
	if config.Format != "table" {
		t.Errorf("Expected format to remain 'table', got %s", config.Format)
	}
	if len(config.Ignore) != 1 || config.Ignore[0] != "*/test/*" {
		t.Errorf("Expected ignore patterns to remain unchanged, got %v", config.Ignore)
	}
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub", "directory")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create config file in parent directory
	configFile := filepath.Join(tempDir, ".gocov.yml")
	if err := os.WriteFile(configFile, []byte("level: 1"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to subdirectory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test finding config file
	found := FindConfigFile()
	if found == "" {
		t.Error("Expected to find config file in parent directory")
	}

	// Verify the found file is correct
	if filepath.Base(found) != ".gocov.yml" {
		t.Errorf("Expected to find .gocov.yml, got %s", filepath.Base(found))
	}
}

