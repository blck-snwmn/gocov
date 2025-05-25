package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/cover"
)

type dirCoverage struct {
	dir         string
	stmtCount   int
	stmtCovered int
}

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
	var coverProfile string
	var level int
	var minCoverage, maxCoverage float64
	var outputFormat string
	var ignoreDirs string
	var configFile string
	flag.StringVar(&coverProfile, "coverprofile", "", "Path to coverage profile file")
	flag.IntVar(&level, "level", 0, "Directory level for aggregation (0 for leaf directories, -1 for all levels)")
	flag.Float64Var(&minCoverage, "min", 0.0, "Minimum coverage percentage to display (0-100)")
	flag.Float64Var(&maxCoverage, "max", 100.0, "Maximum coverage percentage to display (0-100)")
	flag.StringVar(&outputFormat, "format", "table", "Output format (table or json)")
	flag.StringVar(&ignoreDirs, "ignore", "", "Comma-separated list of directories to ignore (supports wildcards)")
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config := DefaultConfig()

	// Try to find config file if not specified
	if configFile == "" {
		configFile = FindConfigFile()
	}

	// Load config file if found
	if configFile != "" {
		loadedConfig, err := LoadConfig(configFile)
		if err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
		if loadedConfig != nil {
			config = loadedConfig
		}
	}

	// Parse ignore patterns from command line
	var ignorePatterns []string
	if ignoreDirs != "" {
		ignorePatterns = strings.Split(ignoreDirs, ",")
		for i := range ignorePatterns {
			ignorePatterns[i] = strings.TrimSpace(ignorePatterns[i])
		}
	}

	// Merge command line flags with config
	config.MergeWithFlags(&level, &minCoverage, &maxCoverage, &outputFormat, ignorePatterns)

	// Apply final configuration values
	if coverProfile == "" {
		flag.Usage()
		os.Exit(1)
	}

	level = config.Level
	minCoverage = config.Coverage.Min
	maxCoverage = config.Coverage.Max
	outputFormat = config.Format
	ignorePatterns = config.Ignore

	if minCoverage < 0 || minCoverage > 100 {
		log.Fatalf("min must be between 0 and 100")
	}
	if maxCoverage < 0 || maxCoverage > 100 {
		log.Fatalf("max must be between 0 and 100")
	}
	if minCoverage > maxCoverage {
		log.Fatalf("min cannot be greater than max")
	}

	profiles, err := cover.ParseProfiles(coverProfile)
	if err != nil {
		log.Fatalf("Failed to parse coverage profile: %v", err)
	}

	coverageByDir := aggregateCoverageByDirectory(profiles, level, ignorePatterns)

	// Select formatter based on output format
	var formatter OutputFormatter
	switch outputFormat {
	case "json":
		formatter = &JSONFormatter{writer: os.Stdout}
	case "table":
		formatter = &TableFormatter{writer: os.Stdout}
	default:
		log.Fatalf("Unknown output format: %s", outputFormat)
	}

	if err := displayResultsWithFormatter(coverageByDir, minCoverage, maxCoverage, formatter); err != nil {
		log.Fatalf("Failed to display results: %v", err)
	}
}

func aggregateCoverageByDirectory(profiles []*cover.Profile, level int, ignoredPatterns []string) map[string]*dirCoverage {
	coverageByDir := make(map[string]*dirCoverage)

	for _, profile := range profiles {
		dir := filepath.Dir(profile.FileName)

		// Check if directory should be ignored
		if shouldIgnoreDirectory(dir, ignoredPatterns) {
			continue
		}

		// Adjust directory path based on level
		if level > 0 {
			parts := strings.Split(dir, string(filepath.Separator))
			if len(parts) > level {
				dir = filepath.Join(parts[:level]...)
			}
		} else if level == -1 {
			// -1 means aggregate at top level (module root)
			dir = "."
		}
		// level == 0 means use the original directory (leaf level)

		if _, exists := coverageByDir[dir]; !exists {
			coverageByDir[dir] = &dirCoverage{dir: dir}
		}

		for _, block := range profile.Blocks {
			stmtCount := block.NumStmt
			coverageByDir[dir].stmtCount += stmtCount

			if block.Count > 0 {
				coverageByDir[dir].stmtCovered += stmtCount
			}
		}
	}

	return coverageByDir
}

// shouldIgnoreDirectory checks if a directory matches any of the ignore patterns
func shouldIgnoreDirectory(dir string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		// Direct match
		matched, err := filepath.Match(pattern, dir)
		if err != nil {
			continue
		}
		if matched {
			return true
		}

		// Check if the pattern contains path separators
		if strings.Contains(pattern, "/") {
			// Try to match against the full path
			if matched, _ := filepath.Match(pattern, dir); matched {
				return true
			}

			// Check if dir contains the pattern
			if strings.Contains(dir, strings.Trim(pattern, "*")) {
				return true
			}
		}

		// Check each component of the path
		parts := strings.Split(dir, string(filepath.Separator))
		for i := range parts {
			// Try matching individual parts
			if matched, _ := filepath.Match(pattern, parts[i]); matched {
				return true
			}

			// Also check combinations from the beginning
			partialPath := filepath.Join(parts[:i+1]...)
			if matched, _ := filepath.Match(pattern, partialPath); matched {
				return true
			}
		}
	}
	return false
}

// calculateCoverage calculates the coverage percentage
func calculateCoverage(stmtCount, stmtCovered int) float64 {
	if stmtCount > 0 {
		return float64(stmtCovered) / float64(stmtCount) * 100
	}
	return 0.0
}

// filterDirectories filters directories based on coverage thresholds
func filterDirectories(coverageByDir map[string]*dirCoverage, minCoverage, maxCoverage float64) []string {
	var filtered []string
	for dir, cov := range coverageByDir {
		coverage := calculateCoverage(cov.stmtCount, cov.stmtCovered)
		if coverage >= minCoverage && coverage <= maxCoverage {
			filtered = append(filtered, dir)
		}
	}
	sort.Strings(filtered)
	return filtered
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

func displayResultsWithFormatter(coverageByDir map[string]*dirCoverage, minCoverage, maxCoverage float64, formatter OutputFormatter) error {
	// Filter directories based on coverage
	filteredDirs := filterDirectories(coverageByDir, minCoverage, maxCoverage)

	// Build results
	var results []CoverageResult
	filteredStmts := 0
	filteredCovered := 0

	for _, dir := range filteredDirs {
		cov := coverageByDir[dir]
		coverage := calculateCoverage(cov.stmtCount, cov.stmtCovered)

		results = append(results, CoverageResult{
			Directory:  dir,
			Statements: cov.stmtCount,
			Covered:    cov.stmtCovered,
			Coverage:   coverage,
		})

		filteredStmts += cov.stmtCount
		filteredCovered += cov.stmtCovered
	}

	// Calculate totals
	totalStmts := 0
	totalCovered := 0
	for _, cov := range coverageByDir {
		totalStmts += cov.stmtCount
		totalCovered += cov.stmtCovered
	}

	totalResult := CoverageResult{
		Directory:  "TOTAL",
		Statements: totalStmts,
		Covered:    totalCovered,
		Coverage:   calculateCoverage(totalStmts, totalCovered),
	}

	// Prepare filtered total if filters are applied
	var filteredTotal *CoverageResult
	if minCoverage > 0.0 || maxCoverage < 100.0 {
		filteredTotal = &CoverageResult{
			Directory:  "FILTERED TOTAL",
			Statements: filteredStmts,
			Covered:    filteredCovered,
			Coverage:   calculateCoverage(filteredStmts, filteredCovered),
		}
	}

	return formatter.Format(results, totalResult, filteredTotal)
}

// Legacy function for backward compatibility
func displayResults(coverageByDir map[string]*dirCoverage, minCoverage, maxCoverage float64) {
	formatter := &TableFormatter{writer: os.Stdout}
	_ = displayResultsWithFormatter(coverageByDir, minCoverage, maxCoverage, formatter)
}
