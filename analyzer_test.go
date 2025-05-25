package main

import (
	"testing"

	"golang.org/x/tools/cover"
)

func TestNewCoverageAnalyzer(t *testing.T) {
	tests := []struct {
		name           string
		level          int
		ignorePatterns []string
	}{
		{
			name:           "default analyzer",
			level:          0,
			ignorePatterns: nil,
		},
		{
			name:           "with level",
			level:          3,
			ignorePatterns: nil,
		},
		{
			name:           "with ignore patterns",
			level:          0,
			ignorePatterns: []string{"*/test/*", "*/vendor/*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCoverageAnalyzer(tt.level, tt.ignorePatterns)
			if analyzer == nil {
				t.Fatal("NewCoverageAnalyzer returned nil")
			}
			if analyzer.level != tt.level {
				t.Errorf("level = %d, want %d", analyzer.level, tt.level)
			}
			if len(analyzer.ignorePatterns) != len(tt.ignorePatterns) {
				t.Errorf("ignorePatterns length = %d, want %d", len(analyzer.ignorePatterns), len(tt.ignorePatterns))
			}
		})
	}
}

func TestAggregateCoverageByDirectory(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Fatalf("Failed to parse test coverage file: %v", err)
	}

	t.Run("level 0 (leaf directories)", func(t *testing.T) {
		analyzer := NewCoverageAnalyzer(0, nil)
		result := analyzer.Aggregate(profiles)

		tests := []struct {
			dir          string
			wantStmts    int
			wantCovered  int
			wantCoverage float64
		}{
			{
				dir:          "github.com/example/project/pkg/util",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/cmd/server",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/internal/service",
				wantStmts:    7,
				wantCovered:  6,
				wantCoverage: 85.7,
			},
		}

		for _, tt := range tests {
			t.Run(tt.dir, func(t *testing.T) {
				cov, exists := result[tt.dir]
				if !exists {
					t.Fatalf("Directory %s not found in results", tt.dir)
				}

				if cov.StmtCount != tt.wantStmts {
					t.Errorf("StmtCount = %d, want %d", cov.StmtCount, tt.wantStmts)
				}

				if cov.StmtCovered != tt.wantCovered {
					t.Errorf("StmtCovered = %d, want %d", cov.StmtCovered, tt.wantCovered)
				}

				coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
				if coverage < tt.wantCoverage-0.1 || coverage > tt.wantCoverage+0.1 {
					t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, tt.wantCoverage)
				}
			})
		}
	})

	t.Run("level -1 (top level)", func(t *testing.T) {
		analyzer := NewCoverageAnalyzer(-1, nil)
		result := analyzer.Aggregate(profiles)

		cov, exists := result["."]
		if !exists {
			t.Fatal("Top level directory '.' not found in results")
		}

		wantStmts := 21
		wantCovered := 16
		wantCoverage := 76.2

		if cov.StmtCount != wantStmts {
			t.Errorf("StmtCount = %d, want %d", cov.StmtCount, wantStmts)
		}

		if cov.StmtCovered != wantCovered {
			t.Errorf("StmtCovered = %d, want %d", cov.StmtCovered, wantCovered)
		}

		coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
		if coverage < wantCoverage-0.1 || coverage > wantCoverage+0.1 {
			t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, wantCoverage)
		}
	})

	t.Run("level 3 (github.com/example/project)", func(t *testing.T) {
		analyzer := NewCoverageAnalyzer(3, nil)
		result := analyzer.Aggregate(profiles)

		cov, exists := result["github.com/example/project"]
		if !exists {
			t.Fatal("Directory 'github.com/example/project' not found in results")
		}

		wantStmts := 21
		wantCovered := 16
		wantCoverage := 76.2

		if cov.StmtCount != wantStmts {
			t.Errorf("StmtCount = %d, want %d", cov.StmtCount, wantStmts)
		}

		if cov.StmtCovered != wantCovered {
			t.Errorf("StmtCovered = %d, want %d", cov.StmtCovered, wantCovered)
		}

		coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
		if coverage < wantCoverage-0.1 || coverage > wantCoverage+0.1 {
			t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, wantCoverage)
		}
	})

	t.Run("level 4 (github.com/example/project/pkg, cmd, internal)", func(t *testing.T) {
		analyzer := NewCoverageAnalyzer(4, nil)
		result := analyzer.Aggregate(profiles)

		tests := []struct {
			dir          string
			wantStmts    int
			wantCovered  int
			wantCoverage float64
		}{
			{
				dir:          "github.com/example/project/pkg",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/cmd",
				wantStmts:    7,
				wantCovered:  5,
				wantCoverage: 71.4,
			},
			{
				dir:          "github.com/example/project/internal",
				wantStmts:    7,
				wantCovered:  6,
				wantCoverage: 85.7,
			},
		}

		for _, tt := range tests {
			t.Run(tt.dir, func(t *testing.T) {
				cov, exists := result[tt.dir]
				if !exists {
					t.Fatalf("Directory %s not found in results", tt.dir)
				}

				if cov.StmtCount != tt.wantStmts {
					t.Errorf("StmtCount = %d, want %d", cov.StmtCount, tt.wantStmts)
				}

				if cov.StmtCovered != tt.wantCovered {
					t.Errorf("StmtCovered = %d, want %d", cov.StmtCovered, tt.wantCovered)
				}

				coverage := CalculateCoverage(cov.StmtCount, cov.StmtCovered)
				if coverage < tt.wantCoverage-0.1 || coverage > tt.wantCoverage+0.1 {
					t.Errorf("coverage = %.1f%%, want %.1f%%", coverage, tt.wantCoverage)
				}
			})
		}
	})
}

func TestShouldIgnoreDirectory(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		patterns []string
		want     bool
	}{
		{
			name:     "no patterns",
			dir:      "pkg/util",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "exact match",
			dir:      "pkg/util",
			patterns: []string{"pkg/util"},
			want:     true,
		},
		{
			name:     "wildcard match",
			dir:      "pkg/util",
			patterns: []string{"pkg/*"},
			want:     true,
		},
		{
			name:     "parent directory match",
			dir:      "pkg/util/subdir",
			patterns: []string{"pkg"},
			want:     true,
		},
		{
			name:     "no match",
			dir:      "cmd/server",
			patterns: []string{"pkg/*", "internal/*"},
			want:     false,
		},
		{
			name:     "multiple patterns with match",
			dir:      "internal/api",
			patterns: []string{"pkg/*", "internal/*", "cmd/*"},
			want:     true,
		},
		{
			name:     "empty pattern",
			dir:      "pkg/util",
			patterns: []string{"", "pkg/*"},
			want:     true,
		},
		{
			name:     "complex wildcard",
			dir:      "github.com/example/project/internal/api",
			patterns: []string{"*/internal/*"},
			want:     true,
		},
		{
			name:     "invalid pattern",
			dir:      "pkg/util",
			patterns: []string{"[invalid"},
			want:     false,
		},
		{
			name:     "pattern with path separator no match",
			dir:      "pkg/util",
			patterns: []string{"cmd/server"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIgnoreDirectory(tt.dir, tt.patterns)
			if got != tt.want {
				t.Errorf("ShouldIgnoreDirectory(%q, %v) = %v, want %v",
					tt.dir, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestAggregateCoverageWithIgnoredDirectories(t *testing.T) {
	profiles, err := cover.ParseProfiles("testdata/coverage.out")
	if err != nil {
		t.Fatalf("Failed to parse test coverage file: %v", err)
	}

	t.Run("ignore internal directory", func(t *testing.T) {
		ignoredPatterns := []string{"*/internal/*"}
		analyzer := NewCoverageAnalyzer(0, ignoredPatterns)
		result := analyzer.Aggregate(profiles)

		// Should not contain internal/service
		if _, exists := result["github.com/example/project/internal/service"]; exists {
			t.Error("Internal directory should be ignored")
		}

		// Should contain other directories
		if _, exists := result["github.com/example/project/pkg/util"]; !exists {
			t.Error("pkg/util should not be ignored")
		}
		if _, exists := result["github.com/example/project/cmd/server"]; !exists {
			t.Error("cmd/server should not be ignored")
		}
	})

	t.Run("ignore multiple patterns", func(t *testing.T) {
		ignoredPatterns := []string{"*/pkg/*", "*/cmd/*"}
		analyzer := NewCoverageAnalyzer(0, ignoredPatterns)
		result := analyzer.Aggregate(profiles)

		// Should only contain internal directory
		if len(result) != 1 {
			t.Errorf("Expected only 1 directory (internal), got %d", len(result))
		}

		if _, exists := result["github.com/example/project/internal/service"]; !exists {
			t.Error("internal/service should not be ignored")
		}
	})

	t.Run("ignore all with wildcard", func(t *testing.T) {
		ignoredPatterns := []string{"*"}
		analyzer := NewCoverageAnalyzer(0, ignoredPatterns)
		result := analyzer.Aggregate(profiles)

		if len(result) != 0 {
			t.Errorf("Expected all directories to be ignored, got %d", len(result))
		}
	})
}

func TestCalculateCoverage(t *testing.T) {
	tests := []struct {
		name        string
		stmtCount   int
		stmtCovered int
		want        float64
	}{
		{
			name:        "normal coverage",
			stmtCount:   10,
			stmtCovered: 7,
			want:        70.0,
		},
		{
			name:        "zero statements",
			stmtCount:   0,
			stmtCovered: 0,
			want:        0.0,
		},
		{
			name:        "100% coverage",
			stmtCount:   10,
			stmtCovered: 10,
			want:        100.0,
		},
		{
			name:        "0% coverage",
			stmtCount:   10,
			stmtCovered: 0,
			want:        0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCoverage(tt.stmtCount, tt.stmtCovered)
			if got != tt.want {
				t.Errorf("CalculateCoverage(%d, %d) = %f, want %f",
					tt.stmtCount, tt.stmtCovered, got, tt.want)
			}
		})
	}
}

func TestFilterDirectories(t *testing.T) {
	coverageByDir := map[string]*DirCoverage{
		"exactly50": {
			Dir:         "exactly50",
			StmtCount:   10,
			StmtCovered: 5,
		},
		"high": {
			Dir:         "high",
			StmtCount:   10,
			StmtCovered: 8,
		},
		"low": {
			Dir:         "low",
			StmtCount:   10,
			StmtCovered: 2,
		},
	}

	tests := []struct {
		name        string
		minCoverage float64
		maxCoverage float64
		want        []string
	}{
		{
			name:        "all directories",
			minCoverage: 0.0,
			maxCoverage: 100.0,
			want:        []string{"exactly50", "high", "low"},
		},
		{
			name:        "minimum threshold",
			minCoverage: 50.0,
			maxCoverage: 100.0,
			want:        []string{"exactly50", "high"},
		},
		{
			name:        "maximum threshold",
			minCoverage: 0.0,
			maxCoverage: 50.0,
			want:        []string{"exactly50", "low"},
		},
		{
			name:        "exact match",
			minCoverage: 50.0,
			maxCoverage: 50.0,
			want:        []string{"exactly50"},
		},
		{
			name:        "no match",
			minCoverage: 90.0,
			maxCoverage: 95.0,
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterDirectories(coverageByDir, tt.minCoverage, tt.maxCoverage)
			if len(got) != len(tt.want) {
				t.Errorf("FilterDirectories() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			for i, dir := range got {
				if dir != tt.want[i] {
					t.Errorf("FilterDirectories()[%d] = %s, want %s", i, dir, tt.want[i])
				}
			}
		})
	}
}

func TestAdjustDirectoryLevel(t *testing.T) {
	tests := []struct {
		name  string
		dir   string
		level int
		want  string
	}{
		{
			name:  "level 0 unchanged",
			dir:   "github.com/example/project/pkg/util",
			level: 0,
			want:  "github.com/example/project/pkg/util",
		},
		{
			name:  "level -1 root",
			dir:   "github.com/example/project/pkg/util",
			level: -1,
			want:  ".",
		},
		{
			name:  "level 3",
			dir:   "github.com/example/project/pkg/util",
			level: 3,
			want:  "github.com/example/project",
		},
		{
			name:  "level exceeds depth",
			dir:   "pkg/util",
			level: 5,
			want:  "pkg/util",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &CoverageAnalyzer{level: tt.level}
			got := analyzer.adjustDirectoryLevel(tt.dir)
			if got != tt.want {
				t.Errorf("adjustDirectoryLevel(%s) = %s, want %s", tt.dir, got, tt.want)
			}
		})
	}
}
