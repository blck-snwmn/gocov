package main

import (
	"path/filepath"
	"sync"

	"golang.org/x/tools/cover"
)

// profileResult represents the result of processing a single profile
type profileResult struct {
	coverageByDir map[string]*DirCoverage
}

// AggregateConcurrent aggregates coverage data by directory using concurrent processing
func (a *CoverageAnalyzer) AggregateConcurrent(profiles []*cover.Profile) map[string]*DirCoverage {
	if len(profiles) <= 10 {
		// For small number of profiles, use sequential processing
		return a.Aggregate(profiles)
	}

	// Use worker pool pattern
	numWorkers := 4
	if len(profiles) < numWorkers {
		numWorkers = len(profiles)
	}

	profileChan := make(chan *cover.Profile, len(profiles))
	resultChan := make(chan profileResult, len(profiles))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for profile := range profileChan {
				if profile == nil {
					continue
				}
				result := a.processProfile(profile)
				resultChan <- profileResult{coverageByDir: result}
			}
		}()
	}

	// Send profiles to workers
	for _, profile := range profiles {
		profileChan <- profile
	}
	close(profileChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Merge results
	finalCoverage := make(map[string]*DirCoverage)
	for result := range resultChan {
		for dir, cov := range result.coverageByDir {
			if existing, exists := finalCoverage[dir]; exists {
				existing.StmtCount += cov.StmtCount
				existing.StmtCovered += cov.StmtCovered
			} else {
				finalCoverage[dir] = &DirCoverage{
					Dir:         cov.Dir,
					StmtCount:   cov.StmtCount,
					StmtCovered: cov.StmtCovered,
				}
			}
		}
	}

	return finalCoverage
}

// processProfile processes a single profile and returns coverage by directory
func (a *CoverageAnalyzer) processProfile(profile *cover.Profile) map[string]*DirCoverage {
	coverageByDir := make(map[string]*DirCoverage)

	dir := filepath.Dir(profile.FileName)

	// Check if directory should be ignored
	if ShouldIgnoreDirectory(dir, a.ignorePatterns) {
		return coverageByDir
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

	return coverageByDir
}
