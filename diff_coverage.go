package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

// DiffCoverageResult represents coverage for changed lines
type DiffCoverageResult struct {
	File           string
	TotalLines     int
	CoveredLines   int
	UncoveredLines []int
	Coverage       float64
}

// DiffCoverageSummary represents the overall diff coverage
type DiffCoverageSummary struct {
	Results      []DiffCoverageResult
	TotalLines   int
	CoveredLines int
	Coverage     float64
}

// CalculateDiffCoverage calculates coverage for changed lines
func CalculateDiffCoverage(profiles []*cover.Profile, diff *GitDiff) *DiffCoverageSummary {
	// Group diff lines by file
	fileChanges := make(map[string][]int)
	for _, line := range diff.Lines {
		fileChanges[line.File] = append(fileChanges[line.File], line.LineNum)
	}

	// Create a map for quick profile lookup
	profileMap := make(map[string]*cover.Profile)
	for _, profile := range profiles {
		// Normalize the file path
		normalizedPath := normalizeFilePath(profile.FileName)
		profileMap[normalizedPath] = profile
	}

	var results []DiffCoverageResult
	totalLines := 0
	totalCovered := 0

	// Calculate coverage for each changed file
	for file, changedLines := range fileChanges {
		// Try to find matching profile
		profile := FindMatchingProfile(profiles, file)

		if profile == nil {
			// File not in coverage profile (maybe not tested at all)
			results = append(results, DiffCoverageResult{
				File:           file,
				TotalLines:     len(changedLines),
				CoveredLines:   0,
				UncoveredLines: changedLines,
				Coverage:       0.0,
			})
			totalLines += len(changedLines)
			continue
		}

		// Check coverage for each changed line
		coveredCount := 0
		var uncoveredLines []int

		for _, lineNum := range changedLines {
			if isLineCovered(profile, lineNum) {
				coveredCount++
			} else {
				uncoveredLines = append(uncoveredLines, lineNum)
			}
		}

		coverage := 0.0
		if len(changedLines) > 0 {
			coverage = float64(coveredCount) / float64(len(changedLines)) * 100
		}

		results = append(results, DiffCoverageResult{
			File:           file,
			TotalLines:     len(changedLines),
			CoveredLines:   coveredCount,
			UncoveredLines: uncoveredLines,
			Coverage:       coverage,
		})

		totalLines += len(changedLines)
		totalCovered += coveredCount
	}

	// Calculate overall coverage
	overallCoverage := 0.0
	if totalLines > 0 {
		overallCoverage = float64(totalCovered) / float64(totalLines) * 100
	}

	return &DiffCoverageSummary{
		Results:      results,
		TotalLines:   totalLines,
		CoveredLines: totalCovered,
		Coverage:     overallCoverage,
	}
}

// isLineCovered checks if a specific line is covered
func isLineCovered(profile *cover.Profile, lineNum int) bool {
	for _, block := range profile.Blocks {
		if lineNum >= block.StartLine && lineNum <= block.EndLine {
			return block.Count > 0
		}
	}
	return false
}

// normalizeFilePath normalizes file paths for comparison
func normalizeFilePath(path string) string {
	// Remove leading ./ if present
	path = strings.TrimPrefix(path, "./")

	// Try to make it absolute if it's not already
	if !filepath.IsAbs(path) {
		// For relative paths, we'll keep them as is
		// The coverage profile might have full package paths
		return path
	}

	return path
}

// FindMatchingProfile tries to find a profile that matches the given file
func FindMatchingProfile(profiles []*cover.Profile, file string) *cover.Profile {
	// Direct match
	for _, profile := range profiles {
		if profile.FileName == file {
			return profile
		}
	}

	// Try with different path combinations
	for _, profile := range profiles {
		// Check if the profile filename ends with our file
		if strings.HasSuffix(profile.FileName, file) {
			return profile
		}

		// Check if the profile filename ends with /file
		if strings.HasSuffix(profile.FileName, "/"+file) {
			return profile
		}

		// Check if our file ends with the profile filename
		if strings.HasSuffix(file, filepath.Base(profile.FileName)) {
			return profile
		}
	}

	return nil
}

// FormatDiffCoverage formats the diff coverage results for display
func FormatDiffCoverage(summary *DiffCoverageSummary) string {
	var output strings.Builder

	output.WriteString("Diff Coverage Report:\n")
	output.WriteString(strings.Repeat("=", 80) + "\n")
	output.WriteString(fmt.Sprintf("%-50s %10s %10s %8s\n", "File", "Lines", "Covered", "Coverage"))
	output.WriteString(strings.Repeat("-", 80) + "\n")

	for _, result := range summary.Results {
		output.WriteString(fmt.Sprintf("%-50s %10d %10d %7.1f%%\n",
			truncateString(result.File, 50),
			result.TotalLines,
			result.CoveredLines,
			result.Coverage))

		// Show uncovered lines if any
		if len(result.UncoveredLines) > 0 && len(result.UncoveredLines) <= 10 {
			uncoveredStr := fmt.Sprintf("  Uncovered lines: %v", result.UncoveredLines)
			output.WriteString(uncoveredStr + "\n")
		} else if len(result.UncoveredLines) > 10 {
			output.WriteString(fmt.Sprintf("  Uncovered lines: %v... (%d more)\n",
				result.UncoveredLines[:10], len(result.UncoveredLines)-10))
		}
	}

	output.WriteString(strings.Repeat("-", 80) + "\n")
	output.WriteString(fmt.Sprintf("%-50s %10d %10d %7.1f%%\n",
		"TOTAL DIFF",
		summary.TotalLines,
		summary.CoveredLines,
		summary.Coverage))

	return output.String()
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
