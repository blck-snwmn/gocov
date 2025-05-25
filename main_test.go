package main

import (
	"bytes"
	"encoding/json"
	"strings"
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

func TestOutputFormatters(t *testing.T) {
	results := []CoverageResult{
		{
			Directory:  "cmd/server",
			Statements: 20,
			Covered:    10,
			Coverage:   50.0,
		},
		{
			Directory:  "pkg/util",
			Statements: 10,
			Covered:    8,
			Coverage:   80.0,
		},
	}

	totalResult := CoverageResult{
		Directory:  "TOTAL",
		Statements: 30,
		Covered:    18,
		Coverage:   60.0,
	}

	t.Run("TableFormatter", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{writer: &buf}

		err := formatter.Format(results, totalResult, nil)
		if err != nil {
			t.Fatalf("TableFormatter failed: %v", err)
		}

		output := buf.String()

		// Verify table output
		if !strings.Contains(output, "Directory") {
			t.Error("Table output should contain header")
		}
		if !strings.Contains(output, "80.0%") {
			t.Error("Table output should contain coverage percentage")
		}
		if !strings.Contains(output, "TOTAL") {
			t.Error("Table output should contain TOTAL line")
		}
	})

	t.Run("TableFormatter with filtered total", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &TableFormatter{writer: &buf}

		filteredTotal := &CoverageResult{
			Directory:  "FILTERED TOTAL",
			Statements: 10,
			Covered:    8,
			Coverage:   80.0,
		}

		err := formatter.Format(results, totalResult, filteredTotal)
		if err != nil {
			t.Fatalf("TableFormatter failed: %v", err)
		}

		output := buf.String()

		// Verify filtered total appears
		if !strings.Contains(output, "FILTERED TOTAL") {
			t.Error("Table output should contain FILTERED TOTAL line")
		}
		if !strings.Contains(output, "TOTAL") {
			t.Error("Table output should still contain TOTAL line")
		}
	})

	t.Run("JSONFormatter", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &JSONFormatter{writer: &buf}

		err := formatter.Format(results, totalResult, nil)
		if err != nil {
			t.Fatalf("JSONFormatter failed: %v", err)
		}

		// Parse JSON output
		var output struct {
			Results []CoverageResult `json:"results"`
			Total   CoverageResult   `json:"total"`
		}

		if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		// Verify JSON structure
		if len(output.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(output.Results))
		}

		if output.Total.Statements != 30 {
			t.Errorf("Expected total statements 30, got %d", output.Total.Statements)
		}

		if output.Total.Covered != 18 {
			t.Errorf("Expected total covered 18, got %d", output.Total.Covered)
		}
	})

	t.Run("JSONFormatter with filters", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &JSONFormatter{writer: &buf}

		filteredTotal := &CoverageResult{
			Directory:  "FILTERED TOTAL",
			Statements: 10,
			Covered:    8,
			Coverage:   80.0,
		}

		err := formatter.Format(results, totalResult, filteredTotal)
		if err != nil {
			t.Fatalf("JSONFormatter failed: %v", err)
		}

		// Parse JSON output
		var output struct {
			Results       []CoverageResult `json:"results"`
			Total         CoverageResult   `json:"total"`
			FilteredTotal *CoverageResult  `json:"filtered_total"`
		}

		if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		// Verify filtered results
		if len(output.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(output.Results))
		}

		if output.FilteredTotal == nil {
			t.Error("Expected filtered_total to be present")
		} else if output.FilteredTotal.Statements != 10 {
			t.Errorf("Expected filtered total statements 10, got %d", output.FilteredTotal.Statements)
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

	t.Run("display with edge case coverages", func(t *testing.T) {
		coverageByDir := map[string]*DirCoverage{
			"zero": {
				Dir:         "zero",
				StmtCount:   0,
				StmtCovered: 0,
			},
			"perfect": {
				Dir:         "perfect",
				StmtCount:   100,
				StmtCovered: 100,
			},
			"none": {
				Dir:         "none",
				StmtCount:   50,
				StmtCovered: 0,
			},
		}

		// Test with TableFormatter
		var buf bytes.Buffer
		formatter := &TableFormatter{writer: &buf}

		// Build results
		var results []CoverageResult
		totalStmts := 0
		totalCovered := 0

		for _, cov := range coverageByDir {
			coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
			results = append(results, CoverageResult{
				Directory:  cov.Dir,
				Statements: cov.StmtCount,
				Covered:    cov.StmtCovered,
				Coverage:   coverage,
			})
			totalStmts += cov.StmtCount
			totalCovered += cov.StmtCovered
		}

		totalResult := CoverageResult{
			Directory:  "TOTAL",
			Statements: totalStmts,
			Covered:    totalCovered,
			Coverage:   CalculateCoverage(totalStmts, totalCovered),
		}

		err := formatter.Format(results, totalResult, nil)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		output := buf.String()

		// Verify edge cases are displayed correctly
		if !strings.Contains(output, "0.0%") {
			t.Error("Output should contain 0.0% coverage")
		}
		if !strings.Contains(output, "100.0%") {
			t.Error("Output should contain 100.0% coverage")
		}
	})
}

