package main

import (
	"bytes"
	"testing"

	"golang.org/x/tools/cover"
)

func TestMainFunction(t *testing.T) {
	// Test that main function properly initializes CLI
	// This is a simple smoke test to ensure main doesn't panic
	t.Run("main smoke test", func(t *testing.T) {
		// We can't easily test main() directly because it calls log.Fatal
		// but we can verify that the components it uses are properly connected
		cli := NewCLI(bytes.NewBuffer(nil), []string{})
		if cli == nil {
			t.Error("CLI initialization failed")
		}
	})
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
		analyzer := NewCoverageAnalyzer(0, nil)
		coverageByDir := analyzer.Aggregate(profiles)
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

		analyzer := NewCoverageAnalyzer(0, nil)
		coverageByDir := analyzer.Aggregate(profiles)
		if len(coverageByDir) != 1 {
			t.Fatalf("Expected 1 directory, got %d", len(coverageByDir))
		}

		for _, cov := range coverageByDir {
			if cov.StmtCount != 0 {
				t.Errorf("Expected 0 statements, got %d", cov.StmtCount)
			}
			coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
			if coverage != 0.0 {
				t.Errorf("Expected 0%% coverage, got %.1f%%", coverage)
			}
		}
	})
}
