package main

import (
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/cover"
)

// DirCoverage represents coverage information for a directory
type DirCoverage struct {
	Dir         string
	StmtCount   int
	StmtCovered int
}

// CoverageAnalyzer analyzes coverage data
type CoverageAnalyzer struct {
	level          int
	ignorePatterns []string
}

// NewCoverageAnalyzer creates a new CoverageAnalyzer
func NewCoverageAnalyzer(level int, ignorePatterns []string) *CoverageAnalyzer {
	return &CoverageAnalyzer{
		level:          level,
		ignorePatterns: ignorePatterns,
	}
}

// Aggregate aggregates coverage data by directory
func (a *CoverageAnalyzer) Aggregate(profiles []*cover.Profile) map[string]*DirCoverage {
	// Pre-allocate map with estimated capacity based on number of profiles
	// Typically, each profile represents one file, and we might have multiple files per directory
	estimatedDirs := len(profiles) / 3
	if estimatedDirs < 10 {
		estimatedDirs = 10
	}
	coverageByDir := make(map[string]*DirCoverage, estimatedDirs)

	for _, profile := range profiles {
		dir := filepath.Dir(profile.FileName)

		// Check if directory should be ignored
		if ShouldIgnoreDirectory(dir, a.ignorePatterns) {
			continue
		}

		// Adjust directory path based on level
		dir = a.adjustDirectoryLevel(dir)

		if _, exists := coverageByDir[dir]; !exists {
			coverageByDir[dir] = &DirCoverage{Dir: dir}
		}

		for _, block := range profile.Blocks {
			stmtCount := block.NumStmt
			coverageByDir[dir].StmtCount += stmtCount

			if block.Count > 0 {
				coverageByDir[dir].StmtCovered += stmtCount
			}
		}
	}

	return coverageByDir
}

func (a *CoverageAnalyzer) adjustDirectoryLevel(dir string) string {
	if a.level > 0 {
		parts := strings.Split(dir, string(filepath.Separator))
		if len(parts) > a.level {
			dir = filepath.Join(parts[:a.level]...)
		}
	} else if a.level == -1 {
		// -1 means aggregate at top level (module root)
		dir = "."
	}
	// level == 0 means use the original directory (leaf level)
	return dir
}

// ShouldIgnoreDirectory checks if a directory matches any of the ignore patterns
func ShouldIgnoreDirectory(dir string, patterns []string) bool {
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

// CalculateCoverage calculates the coverage percentage
func CalculateCoverage(stmtCount, stmtCovered int) float64 {
	if stmtCount > 0 {
		return float64(stmtCovered) / float64(stmtCount) * 100
	}
	return 0.0
}

// FilterDirectories filters directories based on coverage thresholds
func FilterDirectories(coverageByDir map[string]*DirCoverage, minCoverage, maxCoverage float64) []string {
	// Pre-allocate slice with worst-case capacity (all directories)
	filtered := make([]string, 0, len(coverageByDir))
	for dir, cov := range coverageByDir {
		coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
		if coverage >= minCoverage && coverage <= maxCoverage {
			filtered = append(filtered, dir)
		}
	}
	sort.Strings(filtered)
	return filtered
}
