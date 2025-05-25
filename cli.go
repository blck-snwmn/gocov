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
	)

	flags := flag.NewFlagSet("gocov", flag.ContinueOnError)
	flags.SetOutput(c.Output)

	flags.StringVar(&coverProfile, "coverprofile", "", "Path to coverage profile file")
	flags.IntVar(&level, "level", 0, "Directory level for aggregation (0 for leaf directories, -1 for all levels)")
	flags.Float64Var(&minCoverage, "min", 0.0, "Minimum coverage percentage to display (0-100)")
	flags.Float64Var(&maxCoverage, "max", 100.0, "Maximum coverage percentage to display (0-100)")
	flags.StringVar(&outputFormat, "format", "table", "Output format (table or json)")
	flags.StringVar(&ignoreDirs, "ignore", "", "Comma-separated list of directories to ignore (supports wildcards)")
	flags.StringVar(&configFile, "config", "", "Path to configuration file")

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
	config.MergeWithFlags(&level, &minCoverage, &maxCoverage, &outputFormat, config.Ignore)

	// Validate configuration
	if err := c.validateConfiguration(config); err != nil {
		return err
	}

	// Parse coverage profile
	profiles, err := cover.ParseProfiles(coverProfile)
	if err != nil {
		return NewParseError(coverProfile, err)
	}

	// Create analyzer
	analyzer := NewCoverageAnalyzer(config.Level, config.Ignore)
	coverageByDir := analyzer.Aggregate(profiles)

	// Create formatter
	formatter, err := c.createFormatter(config.Format)
	if err != nil {
		return err
	}

	// Display results
	return c.displayResults(coverageByDir, config.Coverage.Min, config.Coverage.Max, formatter)
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
	return ValidateCoverageConfig(config.Coverage.Min, config.Coverage.Max)
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

func (c *CLI) displayResults(coverageByDir map[string]*DirCoverage, minCoverage, maxCoverage float64, formatter OutputFormatter) error {
	// Filter directories based on coverage
	filteredDirs := FilterDirectories(coverageByDir, minCoverage, maxCoverage)

	// Build results
	var results []CoverageResult
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

	return formatter.Format(results, totalResult, filteredTotal)
}
