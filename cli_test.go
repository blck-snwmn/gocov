package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	var buf bytes.Buffer
	args := []string{"-coverprofile", "testdata/coverage.out"}

	cli := NewCLI(&buf, args)

	if cli == nil {
		t.Fatal("NewCLI returned nil")
	}
	if cli.Output != &buf {
		t.Error("Output writer not set correctly")
	}
	if len(cli.Args) != len(args) {
		t.Errorf("Args length = %d, want %d", len(cli.Args), len(args))
	}
}

func TestCLIRun(t *testing.T) {
	t.Run("missing coverprofile", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error for missing coverprofile")
		}
		if !errors.Is(err, ErrNoInput) {
			t.Errorf("Expected ErrNoInput, got: %v", err)
		}
	})

	t.Run("invalid coverage profile", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{"-coverprofile", "testdata/invalid.out"})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error for invalid coverage profile")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected ParseError, got: %v", err)
		}
	})

	t.Run("successful run with table format", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{"-coverprofile", "testdata/coverage.out"})

		err := cli.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		output := buf.String()
		// Check for expected output elements
		if !strings.Contains(output, "Directory") {
			t.Error("Output should contain header 'Directory'")
		}
		if !strings.Contains(output, "TOTAL") {
			t.Error("Output should contain 'TOTAL' line")
		}
	})

	t.Run("successful run with JSON format", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{"-coverprofile", "testdata/coverage.out", "-format", "json"})

		err := cli.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify JSON output
		var result struct {
			Results []CoverageResult `json:"results"`
			Total   CoverageResult   `json:"total"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		if len(result.Results) == 0 {
			t.Error("Expected results in JSON output")
		}
		if result.Total.Statements == 0 {
			t.Error("Expected total statements in JSON output")
		}
	})

	t.Run("with coverage filters", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-min", "50",
			"-max", "80",
		})

		err := cli.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "FILTERED TOTAL") {
			t.Error("Output should contain 'FILTERED TOTAL' when filters are applied")
		}
	})

	t.Run("with ignore patterns", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-ignore", "*/internal/*,*/vendor/*",
		})

		err := cli.Run()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "internal/service") {
			t.Error("Output should not contain ignored directories")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-format", "invalid",
		})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error for invalid format")
		}
		var configErr *ConfigError
		if !errors.As(err, &configErr) {
			t.Errorf("Expected ConfigError, got: %v", err)
		}
	})

	t.Run("invalid min coverage", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-min", "-10",
		})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error for invalid min coverage")
		}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected ValidationError, got: %v", err)
		}
	})

	t.Run("invalid max coverage", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-max", "110",
		})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error for invalid max coverage")
		}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected ValidationError, got: %v", err)
		}
	})

	t.Run("min greater than max", func(t *testing.T) {
		var buf bytes.Buffer
		cli := NewCLI(&buf, []string{
			"-coverprofile", "testdata/coverage.out",
			"-min", "80",
			"-max", "50",
		})

		err := cli.Run()
		if err == nil {
			t.Error("Expected error when min > max")
		}
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("Expected ValidationError, got: %v", err)
		}
	})
}

func TestCLILoadConfiguration(t *testing.T) {
	t.Run("load config file", func(t *testing.T) {
		cli := NewCLI(io.Discard, []string{})
		config, err := cli.loadConfiguration(".gocov.yml", "")
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		// Config file exists, so it should load values from it
		if config.Level != 0 {
			t.Errorf("Expected level from config file, got %d", config.Level)
		}
	})

	t.Run("load invalid config file", func(t *testing.T) {
		// Create an invalid config file
		tempDir := t.TempDir()
		invalidConfig := filepath.Join(tempDir, "invalid.yml")
		if err := os.WriteFile(invalidConfig, []byte("invalid: yaml: content"), 0644); err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		cli := NewCLI(io.Discard, []string{})
		_, err := cli.loadConfiguration(invalidConfig, "")
		if err == nil {
			t.Error("Expected error for invalid config file")
		}
	})

	t.Run("ignore patterns from command line", func(t *testing.T) {
		cli := NewCLI(io.Discard, []string{})
		config, err := cli.loadConfiguration("", "*/test/*, */vendor/*")
		if err != nil {
			t.Fatalf("Failed to load configuration: %v", err)
		}

		if len(config.Ignore) != 2 {
			t.Errorf("Expected 2 ignore patterns, got %d", len(config.Ignore))
		}
		if config.Ignore[0] != "*/test/*" {
			t.Errorf("Expected first pattern '*/test/*', got %s", config.Ignore[0])
		}
		if config.Ignore[1] != "*/vendor/*" {
			t.Errorf("Expected second pattern '*/vendor/*', got %s", config.Ignore[1])
		}
	})
}

func TestCLIDisplayResults(t *testing.T) {
	coverageByDir := map[string]*DirCoverage{
		"pkg/util": {
			Dir:         "pkg/util",
			StmtCount:   10,
			StmtCovered: 8,
		},
		"cmd/server": {
			Dir:         "cmd/server",
			StmtCount:   20,
			StmtCovered: 10,
		},
		"internal/api": {
			Dir:         "internal/api",
			StmtCount:   15,
			StmtCovered: 5,
		},
	}

	t.Run("no filters", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &TableFormatter{writer: w}
		cli := &CLI{Output: w}
		err := cli.displayResults(coverageByDir, 0.0, 100.0, formatter)
		if err != nil {
			t.Fatalf("displayResults failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Check that output contains expected elements
		if !strings.Contains(output, "Directory") {
			t.Error("Output should contain header 'Directory'")
		}

		if !strings.Contains(output, "cmd/server") {
			t.Error("Output should contain 'cmd/server'")
		}

		if !strings.Contains(output, "pkg/util") {
			t.Error("Output should contain 'pkg/util'")
		}

		if !strings.Contains(output, "internal/api") {
			t.Error("Output should contain 'internal/api'")
		}

		if !strings.Contains(output, "TOTAL") {
			t.Error("Output should contain 'TOTAL' line")
		}

		if !strings.Contains(output, "80.0%") {
			t.Error("Output should contain correct coverage percentage for pkg/util (80.0%)")
		}

		if !strings.Contains(output, "50.0%") {
			t.Error("Output should contain correct coverage percentage for cmd/server (50.0%)")
		}

		if !strings.Contains(output, "33.3%") {
			t.Error("Output should contain correct coverage percentage for internal/api (33.3%)")
		}
	})

	t.Run("min coverage filter", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &TableFormatter{writer: w}
		cli := &CLI{Output: w}
		err := cli.displayResults(coverageByDir, 50.0, 100.0, formatter)
		if err != nil {
			t.Fatalf("displayResults failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Should contain directories with >= 50% coverage
		if !strings.Contains(output, "cmd/server") {
			t.Error("Output should contain 'cmd/server' (50.0%)")
		}

		if !strings.Contains(output, "pkg/util") {
			t.Error("Output should contain 'pkg/util' (80.0%)")
		}

		// Should NOT contain directories with < 50% coverage
		if strings.Contains(output, "internal/api") {
			t.Error("Output should NOT contain 'internal/api' (33.3%)")
		}

		if !strings.Contains(output, "FILTERED TOTAL") {
			t.Error("Output should contain 'FILTERED TOTAL' line")
		}
	})

	t.Run("max coverage filter", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &TableFormatter{writer: w}
		cli := &CLI{Output: w}
		err := cli.displayResults(coverageByDir, 0.0, 60.0, formatter)
		if err != nil {
			t.Fatalf("displayResults failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Should contain directories with <= 60% coverage
		if !strings.Contains(output, "cmd/server") {
			t.Error("Output should contain 'cmd/server' (50.0%)")
		}

		if !strings.Contains(output, "internal/api") {
			t.Error("Output should contain 'internal/api' (33.3%)")
		}

		// Should NOT contain directories with > 60% coverage
		if strings.Contains(output, "pkg/util") {
			t.Error("Output should NOT contain 'pkg/util' (80.0%)")
		}

		if !strings.Contains(output, "FILTERED TOTAL") {
			t.Error("Output should contain 'FILTERED TOTAL' line")
		}
	})

	t.Run("range coverage filter", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		formatter := &TableFormatter{writer: w}
		cli := &CLI{Output: w}
		err := cli.displayResults(coverageByDir, 40.0, 70.0, formatter)
		if err != nil {
			t.Fatalf("displayResults failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Should only contain cmd/server (50.0%)
		if !strings.Contains(output, "cmd/server") {
			t.Error("Output should contain 'cmd/server' (50.0%)")
		}

		// Should NOT contain others
		if strings.Contains(output, "pkg/util") {
			t.Error("Output should NOT contain 'pkg/util' (80.0%)")
		}

		if strings.Contains(output, "internal/api") {
			t.Error("Output should NOT contain 'internal/api' (33.3%)")
		}

		if !strings.Contains(output, "FILTERED TOTAL") {
			t.Error("Output should contain 'FILTERED TOTAL' line")
		}
	})
}
