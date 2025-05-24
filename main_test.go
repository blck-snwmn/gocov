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

	t.Run("level 0 (leaf directories)", func(t *testing.T) {
		result := aggregateCoverageByDirectory(profiles, 0)

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
	})

	t.Run("level -1 (top level)", func(t *testing.T) {
		result := aggregateCoverageByDirectory(profiles, -1)

		cov, exists := result["."]
		if !exists {
			t.Fatal("Top level directory '.' not found in results")
		}

		wantStmts := 21
		wantCovered := 16
		wantCoverage := 76.2

		if cov.stmtCount != wantStmts {
			t.Errorf("stmtCount = %d, want %d", cov.stmtCount, wantStmts)
		}

		if cov.stmtCovered != wantCovered {
			t.Errorf("stmtCovered = %d, want %d", cov.stmtCovered, wantCovered)
		}

		coverage := float64(cov.stmtCovered) / float64(cov.stmtCount) * 100
		if coverage < wantCoverage-0.1 || coverage > wantCoverage+0.1 {
			t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, wantCoverage)
		}
	})

	t.Run("level 3 (github.com/example/project)", func(t *testing.T) {
		result := aggregateCoverageByDirectory(profiles, 3)

		cov, exists := result["github.com/example/project"]
		if !exists {
			t.Fatal("Directory 'github.com/example/project' not found in results")
		}

		wantStmts := 21
		wantCovered := 16
		wantCoverage := 76.2

		if cov.stmtCount != wantStmts {
			t.Errorf("stmtCount = %d, want %d", cov.stmtCount, wantStmts)
		}

		if cov.stmtCovered != wantCovered {
			t.Errorf("stmtCovered = %d, want %d", cov.stmtCovered, wantCovered)
		}

		coverage := float64(cov.stmtCovered) / float64(cov.stmtCount) * 100
		if coverage < wantCoverage-0.1 || coverage > wantCoverage+0.1 {
			t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, wantCoverage)
		}
	})

	t.Run("level 4 (github.com/example/project/pkg, cmd, internal)", func(t *testing.T) {
		result := aggregateCoverageByDirectory(profiles, 4)

		tests := []struct {
			dir          string
			wantStmts    int
			wantCovered  int
			wantCoverage float64
		}{
			{
				dir:          "github.com/example/project/pkg",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/cmd",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/internal",
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
	})
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
		"internal/api": {
			dir:         "internal/api",
			stmtCount:   15,
			stmtCovered: 5,
		},
	}

	t.Run("no filters", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayResults(coverageByDir, 0.0, 100.0)

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

		displayResults(coverageByDir, 50.0, 100.0)

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

		displayResults(coverageByDir, 0.0, 60.0)

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

		displayResults(coverageByDir, 40.0, 70.0)

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

func TestMainFunction(t *testing.T) {
	// Test with no arguments (should exit with error)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"gocov"}

	// Test should trigger usage message and exit
	// We can't test the actual exit behavior in unit tests,
	// but we can verify the function handles missing arguments
}

func TestCalculateCoverage(t *testing.T) {
	tests := []struct {
		name        string
		stmtCount   int
		stmtCovered int
		want        float64
	}{
		{
			name:        "normal coverage",
			stmtCount:   10,
			stmtCovered: 7,
			want:        70.0,
		},
		{
			name:        "zero statements",
			stmtCount:   0,
			stmtCovered: 0,
			want:        0.0,
		},
		{
			name:        "100% coverage",
			stmtCount:   10,
			stmtCovered: 10,
			want:        100.0,
		},
		{
			name:        "0% coverage",
			stmtCount:   10,
			stmtCovered: 0,
			want:        0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCoverage(tt.stmtCount, tt.stmtCovered)
			if got != tt.want {
				t.Errorf("calculateCoverage(%d, %d) = %f, want %f",
					tt.stmtCount, tt.stmtCovered, got, tt.want)
			}
		})
	}
}

func TestErrorCases(t *testing.T) {
	t.Run("invalid coverage profile", func(t *testing.T) {
		_, err := cover.ParseProfiles("testdata/invalid.out")
		if err == nil {
			t.Error("Expected error for invalid coverage profile")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := cover.ParseProfiles("testdata/nonexistent.out")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("empty coverage data", func(t *testing.T) {
		profiles, err := cover.ParseProfiles("testdata/empty.out")
		if err != nil {
			t.Fatalf("Failed to parse empty coverage file: %v", err)
		}

		if len(profiles) != 0 {
			t.Errorf("Expected 0 profiles, got %d", len(profiles))
		}

		// Test with empty data
		coverageByDir := aggregateCoverageByDirectory(profiles, 0)
		if len(coverageByDir) != 0 {
			t.Errorf("Expected empty coverage map, got %d entries", len(coverageByDir))
		}
	})
}

func TestBoundaryValues(t *testing.T) {
	t.Run("zero statements file", func(t *testing.T) {
		profiles, err := cover.ParseProfiles("testdata/zerostmt.out")
		if err != nil {
			t.Fatalf("Failed to parse zero statement file: %v", err)
		}

		coverageByDir := aggregateCoverageByDirectory(profiles, 0)
		if len(coverageByDir) != 1 {
			t.Fatalf("Expected 1 directory, got %d", len(coverageByDir))
		}

		for _, cov := range coverageByDir {
			if cov.stmtCount != 0 {
				t.Errorf("Expected 0 statements, got %d", cov.stmtCount)
			}
			coverage := calculateCoverage(cov.stmtCount, cov.stmtCovered)
			if coverage != 0.0 {
				t.Errorf("Expected 0%% coverage, got %.1f%%", coverage)
			}
		}
	})

	t.Run("display with edge case coverages", func(t *testing.T) {
		coverageByDir := map[string]*dirCoverage{
			"zero": {
				dir:         "zero",
				stmtCount:   0,
				stmtCovered: 0,
			},
			"perfect": {
				dir:         "perfect",
				stmtCount:   100,
				stmtCovered: 100,
			},
			"none": {
				dir:         "none",
				stmtCount:   50,
				stmtCovered: 0,
			},
		}

		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayResults(coverageByDir, 0.0, 100.0)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify edge cases are displayed correctly
		if !strings.Contains(output, "0.0%") {
			t.Error("Output should contain 0.0% coverage")
		}
		if !strings.Contains(output, "100.0%") {
			t.Error("Output should contain 100.0% coverage")
		}
	})

	t.Run("filter edge cases", func(t *testing.T) {
		coverageByDir := map[string]*dirCoverage{
			"exactly50": {
				dir:         "exactly50",
				stmtCount:   10,
				stmtCovered: 5,
			},
		}

		// Test boundary inclusion
		filtered := filterDirectories(coverageByDir, 50.0, 50.0)
		if len(filtered) != 1 {
			t.Errorf("Expected exactly 50%% coverage to be included when min=max=50")
		}

		// Test just below boundary
		filtered = filterDirectories(coverageByDir, 50.1, 100.0)
		if len(filtered) != 0 {
			t.Errorf("Expected exactly 50%% coverage to be excluded when min=50.1")
		}

		// Test just above boundary
		filtered = filterDirectories(coverageByDir, 0.0, 49.9)
		if len(filtered) != 0 {
			t.Errorf("Expected exactly 50%% coverage to be excluded when max=49.9")
		}
	})
}
