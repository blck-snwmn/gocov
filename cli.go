package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"golang.org/x/tools/cover"
)

// CLI represents the command-line interface for gocov
type CLI struct {
	Output io.Writer
	Args   []string
}

// NewCLI creates a new CLI instance
func NewCLI(output io.Writer, args []string) *CLI {
	return &CLI{
		Output: output,
		Args:   args,
	}
}

// Run executes the CLI
func (c *CLI) Run() error {
	var (
		coverProfile string
		level        int
		minCoverage  float64
		maxCoverage  float64
		outputFormat string
		ignoreDirs   string
		configFile   string
		concurrent   bool
		threshold    float64
		diffBase     string
	)

	flags := flag.NewFlagSet("gocov", flag.ContinueOnError)
	flags.SetOutput(c.Output)

	flags.StringVar(&coverProfile, "coverprofile", "", "Path to coverage profile file")
	flags.IntVar(&level, "level", 0, "Directory level for aggregation (0 for leaf directories, -1 for all levels)")
	flags.Float64Var(&minCoverage, "min", 0.0, "Minimum coverage percentage to display (0-100)")
	flags.Float64Var(&maxCoverage, "max", 100.0, "Maximum coverage percentage to display (0-100)")
	flags.StringVar(&outputFormat, "format", "", "Output format (table or json)")
	flags.StringVar(&ignoreDirs, "ignore", "", "Comma-separated list of directories to ignore (supports wildcards)")
	flags.StringVar(&configFile, "config", "", "Path to configuration file")
	flags.BoolVar(&concurrent, "concurrent", false, "Use concurrent processing for large coverage files")
	flags.Float64Var(&threshold, "threshold", 0.0, "Minimum total coverage threshold to pass (0-100)")
	flags.StringVar(&diffBase, "diff", "", "Show coverage for changed lines only (e.g., main, HEAD~1)")

	if err := flags.Parse(c.Args); err != nil {
		return err
	}

	// Validate cover profile
	if coverProfile == "" {
		flags.Usage()
		return ErrNoInput
	}

	// Load configuration
	config, err := c.loadConfiguration(configFile, ignoreDirs)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Merge command line flags with config
	config.MergeWithFlags(&level, &minCoverage, &maxCoverage, &outputFormat, config.Ignore, &concurrent, &threshold)

	// Validate configuration
	if err := c.validateConfiguration(config); err != nil {
		return err
	}

	// Parse coverage profile
	profiles, err := cover.ParseProfiles(coverProfile)
	if err != nil {
		return NewParseError(coverProfile, err)
	}

	// Check if diff mode is enabled
	if diffBase != "" {
		return c.runDiffMode(profiles, diffBase, config.Threshold)
	}

	// Create analyzer
	analyzer := NewCoverageAnalyzer(config.Level, config.Ignore)

	// Aggregate coverage data
	var coverageByDir map[string]*DirCoverage
	if config.Concurrent {
		coverageByDir = analyzer.AggregateConcurrent(profiles)
	} else {
		coverageByDir = analyzer.Aggregate(profiles)
	}

	// Create formatter
	formatter, err := c.createFormatter(config.Format)
	if err != nil {
		return err
	}

	// Display results
	totalCoverage, err := c.displayResults(coverageByDir, config.Coverage.Min, config.Coverage.Max, formatter)
	if err != nil {
		return err
	}

	// Check threshold
	if config.Threshold > 0 && totalCoverage < config.Threshold {
		return NewThresholdError(config.Threshold, totalCoverage)
	}

	return nil
}

func (c *CLI) loadConfiguration(configFile, ignoreDirs string) (*Config, error) {
	config := DefaultConfig()

	// Try to find config file if not specified
	if configFile == "" {
		configFile = FindConfigFile()
	}

	// Load config file if found
	if configFile != "" {
		loadedConfig, err := LoadConfig(configFile)
		if err != nil {
			return nil, err
		}
		if loadedConfig != nil {
			config = loadedConfig
		}
	}

	// Parse ignore patterns from command line
	if ignoreDirs != "" {
		ignorePatterns := strings.Split(ignoreDirs, ",")
		for i := range ignorePatterns {
			ignorePatterns[i] = strings.TrimSpace(ignorePatterns[i])
		}
		config.Ignore = ignorePatterns
	}

	return config, nil
}

func (c *CLI) validateConfiguration(config *Config) error {
	if err := ValidateCoverageConfig(config.Coverage.Min, config.Coverage.Max); err != nil {
		return err
	}
	if err := ValidateThreshold(config.Threshold); err != nil {
		return err
	}
	return nil
}

func (c *CLI) createFormatter(format string) (OutputFormatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{writer: c.Output}, nil
	case "table":
		return &TableFormatter{writer: c.Output}, nil
	default:
		return nil, NewConfigError("format", format, ErrInvalidFormat)
	}
}

func (c *CLI) displayResults(coverageByDir map[string]*DirCoverage, minCoverage, maxCoverage float64, formatter OutputFormatter) (float64, error) {
	// Filter directories based on coverage
	filteredDirs := FilterDirectories(coverageByDir, minCoverage, maxCoverage)

	// Build results
	// Pre-allocate with the size of filtered directories
	results := make([]CoverageResult, 0, len(filteredDirs))
	filteredStmts := 0
	filteredCovered := 0

	for _, dir := range filteredDirs {
		cov := coverageByDir[dir]
		coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)

		results = append(results, CoverageResult{
			Directory:  dir,
			Statements: cov.StmtCount,
			Covered:    cov.StmtCovered,
			Coverage:   coverage,
		})

		filteredStmts += cov.StmtCount
		filteredCovered += cov.StmtCovered
	}

	// Calculate totals
	totalStmts := 0
	totalCovered := 0
	for _, cov := range coverageByDir {
		totalStmts += cov.StmtCount
		totalCovered += cov.StmtCovered
	}

	totalResult := CoverageResult{
		Directory:  "TOTAL",
		Statements: totalStmts,
		Covered:    totalCovered,
		Coverage:   CalculateCoverage(totalStmts, totalCovered),
	}

	// Prepare filtered total if filters are applied
	var filteredTotal *CoverageResult
	if minCoverage > 0.0 || maxCoverage < 100.0 {
		filteredTotal = &CoverageResult{
			Directory:  "FILTERED TOTAL",
			Statements: filteredStmts,
			Covered:    filteredCovered,
			Coverage:   CalculateCoverage(filteredStmts, filteredCovered),
		}
	}

	err := formatter.Format(results, totalResult, filteredTotal)
	return totalResult.Coverage, err
}

// runDiffMode runs coverage analysis for changed lines only
func (c *CLI) runDiffMode(profiles []*cover.Profile, diffBase string, threshold float64) error {
	// Get git diff
	diff, err := GetGitDiffWithContext(diffBase)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Calculate diff coverage
	summary := CalculateDiffCoverage(profiles, diff)

	// Format and display results
	fmt.Fprint(c.Output, FormatDiffCoverage(summary))

	// Check threshold if specified
	if threshold > 0 && summary.Coverage < threshold {
		return NewThresholdError(threshold, summary.Coverage)
	}

	return nil
}
