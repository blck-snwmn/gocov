package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// CoverageResult represents the coverage data for output
type CoverageResult struct {
	Directory  string  `json:"directory"`
	Statements int     `json:"statements"`
	Covered    int     `json:"covered"`
	Coverage   float64 `json:"coverage"`
}

// OutputFormatter interface for different output formats
type OutputFormatter interface {
	Format(results []CoverageResult, totalResult CoverageResult, filteredTotal *CoverageResult) error
}

// TableFormatter formats output as a table
type TableFormatter struct {
	writer io.Writer
}

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	writer io.Writer
}

func main() {
	cli := NewCLI(os.Stdout, os.Args[1:])
	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}

// Format implements OutputFormatter for TableFormatter
func (f *TableFormatter) Format(results []CoverageResult, totalResult CoverageResult, filteredTotal *CoverageResult) error {
	// Display header
	fmt.Fprintf(f.writer, "%-50s %10s %10s %8s\n", "Directory", "Statements", "Covered", "Coverage")
	fmt.Fprintln(f.writer, strings.Repeat("-", 80))

	// Display results
	for _, result := range results {
		fmt.Fprintf(f.writer, "%-50s %10d %10d %7.1f%%\n",
			result.Directory, result.Statements, result.Covered, result.Coverage)
	}

	// Display total
	fmt.Fprintln(f.writer, strings.Repeat("-", 80))

	// Show filtered total if provided
	if filteredTotal != nil {
		fmt.Fprintf(f.writer, "%-50s %10d %10d %7.1f%%\n",
			"FILTERED TOTAL", filteredTotal.Statements, filteredTotal.Covered, filteredTotal.Coverage)
	}

	fmt.Fprintf(f.writer, "%-50s %10d %10d %7.1f%%\n",
		"TOTAL", totalResult.Statements, totalResult.Covered, totalResult.Coverage)

	return nil
}

// Format implements OutputFormatter for JSONFormatter
func (f *JSONFormatter) Format(results []CoverageResult, totalResult CoverageResult, filteredTotal *CoverageResult) error {
	output := struct {
		Results       []CoverageResult `json:"results"`
		Total         CoverageResult   `json:"total"`
		FilteredTotal *CoverageResult  `json:"filtered_total,omitempty"`
	}{
		Results:       results,
		Total:         totalResult,
		FilteredTotal: filteredTotal,
	}

	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
