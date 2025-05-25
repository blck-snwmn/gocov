package main

import (
	"strings"
	"testing"

	"golang.org/x/tools/cover"
)

func TestIsLineCovered(t *testing.T) {
	profile := &cover.Profile{
		FileName: "test.go",
		Mode:     "set",
		Blocks: []cover.ProfileBlock{
			{StartLine: 1, EndLine: 5, Count: 1},   // Lines 1-5 are covered
			{StartLine: 10, EndLine: 15, Count: 0}, // Lines 10-15 are not covered
			{StartLine: 20, EndLine: 25, Count: 2}, // Lines 20-25 are covered
		},
	}

	tests := []struct {
		name     string
		lineNum  int
		expected bool
	}{
		{"covered line at start", 1, true},
		{"covered line in middle", 3, true},
		{"covered line at end", 5, true},
		{"uncovered line", 12, false},
		{"line before coverage", 0, false},
		{"line after coverage", 30, false},
		{"line between blocks", 8, false},
		{"covered line with count > 1", 22, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLineCovered(profile, tt.lineNum)
			if got != tt.expected {
				t.Errorf("isLineCovered(profile, %d) = %v, want %v", tt.lineNum, got, tt.expected)
			}
		})
	}
}

func TestFindMatchingProfile(t *testing.T) {
	profiles := []*cover.Profile{
		{FileName: "main.go"},
		{FileName: "internal/service/user.go"},
		{FileName: "github.com/example/project/cmd/main.go"},
		{FileName: "/home/user/go/src/github.com/example/project/pkg/util.go"},
	}

	tests := []struct {
		name     string
		file     string
		wantFile string
		found    bool
	}{
		{
			name:     "exact match",
			file:     "main.go",
			wantFile: "main.go",
			found:    true,
		},
		{
			name:     "suffix match",
			file:     "cmd/main.go",
			wantFile: "github.com/example/project/cmd/main.go",
			found:    true,
		},
		{
			name:     "suffix with slash",
			file:     "service/user.go",
			wantFile: "internal/service/user.go",
			found:    true,
		},
		{
			name:     "no match",
			file:     "notfound.go",
			wantFile: "",
			found:    false,
		},
		{
			name:     "partial path match",
			file:     "pkg/util.go",
			wantFile: "/home/user/go/src/github.com/example/project/pkg/util.go",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindMatchingProfile(profiles, tt.file)
			if tt.found {
				if got == nil {
					t.Errorf("FindMatchingProfile() = nil, want profile with FileName %q", tt.wantFile)
				} else if got.FileName != tt.wantFile {
					t.Errorf("FindMatchingProfile() returned profile with FileName %q, want %q", got.FileName, tt.wantFile)
				}
			} else {
				if got != nil {
					t.Errorf("FindMatchingProfile() = %v, want nil", got)
				}
			}
		})
	}
}

func TestCalculateDiffCoverage(t *testing.T) {
	// Create test profiles
	profiles := []*cover.Profile{
		{
			FileName: "main.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 10, EndLine: 20, Count: 1},
				{StartLine: 30, EndLine: 40, Count: 0},
			},
		},
		{
			FileName: "internal/service/user.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 5, EndLine: 15, Count: 1},
				{StartLine: 20, EndLine: 25, Count: 0},
			},
		},
	}

	// Create test diff
	diff := &GitDiff{
		BaseRef: "HEAD~1",
		Lines: []DiffLine{
			{File: "main.go", LineNum: 15, ChangeType: "added"},         // covered
			{File: "main.go", LineNum: 35, ChangeType: "added"},         // not covered
			{File: "service/user.go", LineNum: 10, ChangeType: "added"}, // covered
			{File: "service/user.go", LineNum: 22, ChangeType: "added"}, // not covered
			{File: "newfile.go", LineNum: 1, ChangeType: "added"},       // file not in coverage
		},
	}

	summary := CalculateDiffCoverage(profiles, diff)

	// Verify total statistics
	if summary.TotalLines != 5 {
		t.Errorf("TotalLines = %d, want 5", summary.TotalLines)
	}
	if summary.CoveredLines != 2 {
		t.Errorf("CoveredLines = %d, want 2", summary.CoveredLines)
	}
	if summary.Coverage != 40.0 {
		t.Errorf("Coverage = %.1f%%, want 40.0%%", summary.Coverage)
	}

	// Verify we have results for all files
	if len(summary.Results) != 3 {
		t.Errorf("Results count = %d, want 3", len(summary.Results))
	}

	// Check specific file results
	for _, result := range summary.Results {
		switch result.File {
		case "main.go":
			if result.TotalLines != 2 || result.CoveredLines != 1 {
				t.Errorf("main.go: got %d/%d lines, want 2/1", result.CoveredLines, result.TotalLines)
			}
		case "service/user.go":
			if result.TotalLines != 2 || result.CoveredLines != 1 {
				t.Errorf("service/user.go: got %d/%d lines, want 2/1", result.CoveredLines, result.TotalLines)
			}
		case "newfile.go":
			if result.TotalLines != 1 || result.CoveredLines != 0 {
				t.Errorf("newfile.go: got %d/%d lines, want 1/0", result.CoveredLines, result.TotalLines)
			}
		}
	}
}

func TestFormatDiffCoverage(t *testing.T) {
	summary := &DiffCoverageSummary{
		Results: []DiffCoverageResult{
			{
				File:           "main.go",
				TotalLines:     10,
				CoveredLines:   8,
				UncoveredLines: []int{15, 16},
				Coverage:       80.0,
			},
			{
				File:           "very/long/path/to/file/that/should/be/truncated/service.go",
				TotalLines:     5,
				CoveredLines:   0,
				UncoveredLines: []int{1, 2, 3, 4, 5},
				Coverage:       0.0,
			},
			{
				File:           "pkg/util.go",
				TotalLines:     20,
				CoveredLines:   20,
				UncoveredLines: []int{},
				Coverage:       100.0,
			},
		},
		TotalLines:   35,
		CoveredLines: 28,
		Coverage:     80.0,
	}

	output := FormatDiffCoverage(summary)

	// Check that output contains expected elements
	expectedStrings := []string{
		"Diff Coverage Report:",
		"main.go",
		"80.0%",
		"Uncovered lines: [15 16]",
		"very/long/path/to/file/that/should/be/truncated...",
		"0.0%",
		"pkg/util.go",
		"100.0%",
		"TOTAL DIFF",
		"35",
		"28",
		"80.0%",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("FormatDiffCoverage() output missing expected string: %q", expected)
		}
	}

	// Check that long uncovered lines list is truncated
	manyUncovered := &DiffCoverageSummary{
		Results: []DiffCoverageResult{
			{
				File:           "test.go",
				TotalLines:     20,
				CoveredLines:   5,
				UncoveredLines: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				Coverage:       25.0,
			},
		},
		TotalLines:   20,
		CoveredLines: 5,
		Coverage:     25.0,
	}

	output2 := FormatDiffCoverage(manyUncovered)
	if !strings.Contains(output2, "... (5 more)") {
		t.Error("FormatDiffCoverage() should truncate long uncovered lines list")
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "needs truncation",
			input:  "this is a very long string",
			maxLen: 10,
			want:   "this is...",
		},
		{
			name:   "very short max",
			input:  "hello",
			maxLen: 3,
			want:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
