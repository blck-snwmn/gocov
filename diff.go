package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DiffLine represents a changed line in a file
type DiffLine struct {
	File       string
	LineNum    int
	ChangeType string // "added" or "modified"
}

// GitDiff represents the diff information
type GitDiff struct {
	BaseRef string
	Lines   []DiffLine
}

// GetGitDiff retrieves the diff between the base reference and HEAD
func GetGitDiff(baseRef string) (*GitDiff, error) {
	if baseRef == "" {
		baseRef = "HEAD~1"
	}

	var cmd *exec.Cmd
	switch baseRef {
	case "staged", "cached":
		// Get staged changes
		cmd = exec.Command("git", "diff", "--cached", "--unified=0")
	case "working", "unstaged":
		// Get unstaged changes
		cmd = exec.Command("git", "diff", "--unified=0")
	default:
		// Get diff between commits
		cmd = exec.Command("git", "diff", baseRef, "HEAD", "--unified=0")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	diff := &GitDiff{
		BaseRef: baseRef,
		Lines:   []DiffLine{},
	}

	// Parse the diff output
	lines := strings.Split(string(output), "\n")
	var currentFile string

	for _, line := range lines {
		// File header
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			continue
		}

		// Hunk header (e.g., @@ -1,3 +1,5 @@)
		if strings.HasPrefix(line, "@@") && currentFile != "" {
			hunkInfo := parseHunkHeader(line)
			if hunkInfo != nil {
				// Process added lines in this hunk
				addedLines := getAddedLinesFromHunk(lines, line, hunkInfo)
				for _, lineNum := range addedLines {
					diff.Lines = append(diff.Lines, DiffLine{
						File:       currentFile,
						LineNum:    lineNum,
						ChangeType: "added",
					})
				}
			}
		}
	}

	return diff, nil
}

// HunkInfo represents the information from a hunk header
type HunkInfo struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
}

// parseHunkHeader parses a hunk header like "@@ -1,3 +1,5 @@"
func parseHunkHeader(header string) *HunkInfo {
	// Check if it's a valid hunk header
	if !strings.HasPrefix(header, "@@") {
		return nil
	}

	// Extract the numbers from the header
	parts := strings.Split(header, " ")
	if len(parts) < 3 {
		return nil
	}

	// Parse old file info
	oldInfo := strings.TrimPrefix(parts[1], "-")
	oldParts := strings.Split(oldInfo, ",")
	oldStart, _ := strconv.Atoi(oldParts[0])
	oldCount := 1
	if len(oldParts) > 1 {
		oldCount, _ = strconv.Atoi(oldParts[1])
	}

	// Parse new file info
	newInfo := strings.TrimPrefix(parts[2], "+")
	newParts := strings.Split(newInfo, ",")
	newStart, _ := strconv.Atoi(newParts[0])
	newCount := 1
	if len(newParts) > 1 {
		newCount, _ = strconv.Atoi(newParts[1])
	}

	return &HunkInfo{
		OldStart: oldStart,
		OldCount: oldCount,
		NewStart: newStart,
		NewCount: newCount,
	}
}

// getAddedLinesFromHunk extracts line numbers of added lines from a hunk
func getAddedLinesFromHunk(allLines []string, hunkHeader string, info *HunkInfo) []int {
	addedLines := []int{}

	// Find the hunk header position
	hunkIndex := -1
	for i, line := range allLines {
		if line == hunkHeader {
			hunkIndex = i
			break
		}
	}

	if hunkIndex == -1 {
		return addedLines
	}

	// Process lines after the hunk header
	currentLine := info.NewStart
	linesProcessed := 0

	for i := hunkIndex + 1; i < len(allLines) && linesProcessed < info.NewCount; i++ {
		line := allLines[i]
		if strings.HasPrefix(line, "@@") {
			break // Next hunk
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, currentLine)
			currentLine++
			linesProcessed++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// Deleted lines don't count towards new file line numbers
		} else if !strings.HasPrefix(line, "\\") {
			// Context line
			currentLine++
			linesProcessed++
		}
	}

	return addedLines
}

// GetGitDiffWithContext gets diff with more sophisticated parsing
func GetGitDiffWithContext(baseRef string) (*GitDiff, error) {
	if baseRef == "" {
		// Try to find the merge base with main/master
		mergeBase, err := getMergeBase()
		if err == nil {
			baseRef = mergeBase
		} else {
			baseRef = "HEAD~1"
		}
	}

	// Use git diff with name-status to get changed files first
	var cmd *exec.Cmd
	switch baseRef {
	case "staged", "cached":
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
	case "working", "unstaged":
		cmd = exec.Command("git", "diff", "--name-only")
	default:
		cmd = exec.Command("git", "diff", baseRef, "HEAD", "--name-only")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	diff := &GitDiff{
		BaseRef: baseRef,
		Lines:   []DiffLine{},
	}

	// For each changed file, get detailed line changes
	for _, file := range changedFiles {
		if file == "" || !strings.HasSuffix(file, ".go") {
			continue
		}

		// Get diff for specific file
		switch baseRef {
		case "staged", "cached":
			cmd = exec.Command("git", "diff", "--cached", "--", file)
		case "working", "unstaged":
			cmd = exec.Command("git", "diff", "--", file)
		default:
			cmd = exec.Command("git", "diff", baseRef, "HEAD", "--", file)
		}

		fileDiff, err := cmd.Output()
		if err != nil {
			continue
		}

		// Parse the file diff
		lines := parseFileDiff(file, string(fileDiff))
		diff.Lines = append(diff.Lines, lines...)
	}

	return diff, nil
}

// parseFileDiff parses diff output for a single file
func parseFileDiff(filename string, diffContent string) []DiffLine {
	var result []DiffLine
	scanner := bufio.NewScanner(strings.NewReader(diffContent))

	var currentNewLine int
	inHunk := false

	for scanner.Scan() {
		line := scanner.Text()

		// Hunk header
		if strings.HasPrefix(line, "@@") {
			info := parseHunkHeader(line)
			if info != nil {
				currentNewLine = info.NewStart
				inHunk = true
			}
			continue
		}

		if !inHunk {
			continue
		}

		// Added line
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			result = append(result, DiffLine{
				File:       filename,
				LineNum:    currentNewLine,
				ChangeType: "added",
			})
			currentNewLine++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// Deleted line, don't increment line number
		} else if !strings.HasPrefix(line, "\\") {
			// Context line
			currentNewLine++
		}
	}

	return result
}

// getMergeBase tries to find the merge base with main or master branch
func getMergeBase() (string, error) {
	// Try main branch first
	cmd := exec.Command("git", "merge-base", "HEAD", "main")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// Try master branch
	cmd = exec.Command("git", "merge-base", "HEAD", "master")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	return "", fmt.Errorf("could not find merge base")
}
