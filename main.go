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
	flag.StringVar(&coverProfile, "coverprofile", "", "Path to coverage profile file")
	flag.Parse()

	if coverProfile == "" {
		flag.Usage()
		os.Exit(1)
	}

	profiles, err := cover.ParseProfiles(coverProfile)
	if err != nil {
		log.Fatalf("Failed to parse coverage profile: %v", err)
	}

	coverageByDir := aggregateCoverageByDirectory(profiles)
	displayResults(coverageByDir)
}

func aggregateCoverageByDirectory(profiles []*cover.Profile) map[string]*dirCoverage {
	coverageByDir := make(map[string]*dirCoverage)

	for _, profile := range profiles {
		dir := filepath.Dir(profile.FileName)
		
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

func displayResults(coverageByDir map[string]*dirCoverage) {
	// Sort directories for consistent output
	var dirs []string
	for dir := range coverageByDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// Display header
	fmt.Printf("%-50s %10s %10s %8s\n", "Directory", "Statements", "Covered", "Coverage")
	fmt.Println(strings.Repeat("-", 80))

	// Display coverage for each directory
	totalStmts := 0
	totalCovered := 0
	
	for _, dir := range dirs {
		cov := coverageByDir[dir]
		coverage := 0.0
		if cov.stmtCount > 0 {
			coverage = float64(cov.stmtCovered) / float64(cov.stmtCount) * 100
		}
		
		fmt.Printf("%-50s %10d %10d %7.1f%%\n", 
			dir, cov.stmtCount, cov.stmtCovered, coverage)
		
		totalStmts += cov.stmtCount
		totalCovered += cov.stmtCovered
	}

	// Display total
	fmt.Println(strings.Repeat("-", 80))
	
	totalCoverage := 0.0
	if totalStmts > 0 {
		totalCoverage = float64(totalCovered) / float64(totalStmts) * 100
	}
	
	fmt.Printf("%-50s %10d %10d %7.1f%%\n", 
		"TOTAL", totalStmts, totalCovered, totalCoverage)
}