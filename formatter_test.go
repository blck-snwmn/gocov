package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

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

func TestFormatterEdgeCases(t *testing.T) {
	t.Run("display with edge case coverages", func(t *testing.T) {
		results := []CoverageResult{
			{
				Directory:  "zero",
				Statements: 0,
				Covered:    0,
				Coverage:   0.0,
			},
			{
				Directory:  "perfect",
				Statements: 100,
				Covered:    100,
				Coverage:   100.0,
			},
			{
				Directory:  "none",
				Statements: 50,
				Covered:    0,
				Coverage:   0.0,
			},
		}

		totalResult := CoverageResult{
			Directory:  "TOTAL",
			Statements: 150,
			Covered:    100,
			Coverage:   66.7,
		}

		// Test with TableFormatter
		var buf bytes.Buffer
		formatter := &TableFormatter{writer: &buf}

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
