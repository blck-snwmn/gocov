package main

import (
	"flag"
	"fmt"
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

func main() {
	var coverProfile string
	var level int
	var minCoverage, maxCoverage float64
	flag.StringVar(&coverProfile, "coverprofile", "", "Path to coverage profile file")
	flag.IntVar(&level, "level", 0, "Directory level for aggregation (0 for leaf directories, -1 for all levels)")
	flag.Float64Var(&minCoverage, "min", 0.0, "Minimum coverage percentage to display (0-100)")
	flag.Float64Var(&maxCoverage, "max", 100.0, "Maximum coverage percentage to display (0-100)")
	flag.Parse()

	if coverProfile == "" {
		flag.Usage()
		os.Exit(1)
	}

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

	coverageByDir := aggregateCoverageByDirectory(profiles, level)
	displayResults(coverageByDir, minCoverage, maxCoverage)
}

func aggregateCoverageByDirectory(profiles []*cover.Profile, level int) map[string]*dirCoverage {
	coverageByDir := make(map[string]*dirCoverage)

	for _, profile := range profiles {
		dir := filepath.Dir(profile.FileName)

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

func displayResults(coverageByDir map[string]*dirCoverage, minCoverage, maxCoverage float64) {
	// Get all directories sorted
	var allDirs []string
	for dir := range coverageByDir {
		allDirs = append(allDirs, dir)
	}
	sort.Strings(allDirs)

	// Filter directories based on coverage
	filteredDirs := filterDirectories(coverageByDir, minCoverage, maxCoverage)

	// Display header
	fmt.Printf("%-50s %10s %10s %8s\n", "Directory", "Statements", "Covered", "Coverage")
	fmt.Println(strings.Repeat("-", 80))

	// Display coverage for filtered directories
	filteredStmts := 0
	filteredCovered := 0

	for _, dir := range filteredDirs {
		cov := coverageByDir[dir]
		coverage := calculateCoverage(cov.stmtCount, cov.stmtCovered)

		fmt.Printf("%-50s %10d %10d %7.1f%%\n",
			dir, cov.stmtCount, cov.stmtCovered, coverage)

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

	// Display total
	fmt.Println(strings.Repeat("-", 80))

	// Show filtered total if filters are applied
	if minCoverage > 0.0 || maxCoverage < 100.0 {
		filteredCoverage := calculateCoverage(filteredStmts, filteredCovered)
		fmt.Printf("%-50s %10d %10d %7.1f%%\n",
			"FILTERED TOTAL", filteredStmts, filteredCovered, filteredCoverage)
	}

	totalCoverage := calculateCoverage(totalStmts, totalCovered)
	fmt.Printf("%-50s %10d %10d %7.1f%%\n",
		"TOTAL", totalStmts, totalCovered, totalCoverage)
}
