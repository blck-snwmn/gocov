package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != 0 {
		t.Errorf("Expected default level to be 0, got %d", config.Level)
	}
	if config.Coverage.Min != 0 {
		t.Errorf("Expected default min coverage to be 0, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 100 {
		t.Errorf("Expected default max coverage to be 100, got %f", config.Coverage.Max)
	}
	if config.Format != "table" {
		t.Errorf("Expected default format to be 'table', got %s", config.Format)
	}
	if len(config.Ignore) != 0 {
		t.Errorf("Expected default ignore list to be empty, got %v", config.Ignore)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".gocov.yml")

	configContent := `level: 2
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
		expectError string
	}{
		{
			name: "invalid min coverage below 0",
			configYAML: `coverage:
  min: -10
  max: 100
`,
			expectError: "invalid min coverage",
		},
		{
			name: "invalid min coverage above 100",
			configYAML: `coverage:
  min: 150
  max: 100
`,
			expectError: "invalid min coverage",
		},
		{
			name: "invalid max coverage below 0",
			configYAML: `coverage:
  min: 0
  max: -10
`,
			expectError: "invalid max coverage",
		},
		{
			name: "invalid max coverage above 100",
			configYAML: `coverage:
  min: 0
  max: 150
`,
			expectError: "invalid max coverage",
		},
		{
			name: "min greater than max",
			configYAML: `coverage:
  min: 80
  max: 50
`,
			expectError: "min coverage",
		},
		{
			name:        "invalid format",
			configYAML:  `format: xml`,
			expectError: "invalid format",
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
				t.Errorf("Expected error containing '%s', got nil", tt.expectError)
			} else if !containsString(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectError, err)
			}
		})
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	config, err := LoadConfig("/non/existent/config.yml")
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got %v", err)
	}
	if config != nil {
		t.Errorf("Expected nil config for non-existent file, got %v", config)
	}
}

func TestMergeWithFlags(t *testing.T) {
	config := DefaultConfig()

	// Test merging with flags
	level := 3
	minCov := 60.0
	maxCov := 95.0
	format := "json"
	ignorePatterns := []string{"*/vendor/*", "*/test/*"}

	config.MergeWithFlags(&level, &minCov, &maxCov, &format, ignorePatterns)

	if config.Level != 3 {
		t.Errorf("Expected level to be 3, got %d", config.Level)
	}
	if config.Coverage.Min != 60.0 {
		t.Errorf("Expected min coverage to be 60.0, got %f", config.Coverage.Min)
	}
	if config.Coverage.Max != 95.0 {
		t.Errorf("Expected max coverage to be 95.0, got %f", config.Coverage.Max)
	}
	if config.Format != "json" {
		t.Errorf("Expected format to be 'json', got %s", config.Format)
	}
	if len(config.Ignore) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(config.Ignore))
	}

	// Test merging with nil flags (should keep original values)
	config2 := DefaultConfig()
	config2.MergeWithFlags(nil, nil, nil, nil, nil)

	if config2.Level != 0 {
		t.Errorf("Expected level to remain 0, got %d", config2.Level)
	}
	if config2.Coverage.Min != 0 {
		t.Errorf("Expected min coverage to remain 0, got %f", config2.Coverage.Min)
	}
	if config2.Coverage.Max != 100 {
		t.Errorf("Expected max coverage to remain 100, got %f", config2.Coverage.Max)
	}
	if config2.Format != "table" {
		t.Errorf("Expected format to remain 'table', got %s", config2.Format)
	}
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create config file in parent directory
	configFile := filepath.Join(tempDir, ".gocov.yml")
	if err := os.WriteFile(configFile, []byte("level: 1"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Change to subdirectory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test finding config file in parent directory
	found := FindConfigFile()
	// Compare base names since paths might be symlinked
	if filepath.Base(found) != filepath.Base(configFile) || !fileExists(found) {
		t.Errorf("Expected to find config file, got %s", found)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && contains(s, substr)))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
