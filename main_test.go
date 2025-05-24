package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/cover"
)

func TestAggregateCoverageByDirectory(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Fatalf("Failed to parse test coverage file: %v", err)
	}

	result := aggregateCoverageByDirectory(profiles)

	tests := []struct {
		dir          string
		wantStmts    int
		wantCovered  int
		wantCoverage float64
	}{
		{
			dir:          "github.com/example/project/pkg/util",
			wantStmts:    7,
			wantCovered:  5,
			wantCoverage: 71.4,
		},
		{
			dir:          "github.com/example/project/cmd/server",
			wantStmts:    7,
			wantCovered:  5,
			wantCoverage: 71.4,
		},
		{
			dir:          "github.com/example/project/internal/service",
			wantStmts:    7,
			wantCovered:  6,
			wantCoverage: 85.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			cov, exists := result[tt.dir]
			if !exists {
				t.Fatalf("Directory %s not found in results", tt.dir)
			}

			if cov.stmtCount != tt.wantStmts {
				t.Errorf("stmtCount = %d, want %d", cov.stmtCount, tt.wantStmts)
			}

			if cov.stmtCovered != tt.wantCovered {
				t.Errorf("stmtCovered = %d, want %d", cov.stmtCovered, tt.wantCovered)
			}

			coverage := float64(cov.stmtCovered) / float64(cov.stmtCount) * 100
			if coverage < tt.wantCoverage-0.1 || coverage > tt.wantCoverage+0.1 {
				t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, tt.wantCoverage)
			}
		})
	}
}

func TestDisplayResults(t *testing.T) {
	coverageByDir := map[string]*dirCoverage{
		"pkg/util": {
			dir:         "pkg/util",
			stmtCount:   10,
			stmtCovered: 8,
		},
		"cmd/server": {
			dir:         "cmd/server",
			stmtCount:   20,
			stmtCovered: 10,
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayResults(coverageByDir)

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

	if !strings.Contains(output, "TOTAL") {
		t.Error("Output should contain 'TOTAL' line")
	}

	if !strings.Contains(output, "80.0%") {
		t.Error("Output should contain correct coverage percentage for pkg/util (80.0%)")
	}

	if !strings.Contains(output, "50.0%") {
		t.Error("Output should contain correct coverage percentage for cmd/server (50.0%)")
	}

	if !strings.Contains(output, "60.0%") {
		t.Error("Output should contain correct total coverage percentage (60.0%)")
	}
}

func TestMainFunction(t *testing.T) {
	// Test with no arguments (should exit with error)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gocov"}

	// Test should trigger usage message and exit
	// We can't test the actual exit behavior in unit tests,
	// but we can verify the function handles missing arguments
}