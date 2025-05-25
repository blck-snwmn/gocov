package main

import (
	"reflect"
	"testing"

	"golang.org/x/tools/cover"
)

func TestAggregateConcurrent(t *testing.T) {
	// Create test profiles
	profiles := []*cover.Profile{
		{
			FileName: "github.com/example/project/pkg/util/helper.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
				{StartLine: 11, StartCol: 1, EndLine: 20, EndCol: 1, NumStmt: 5, Count: 0},
			},
		},
		{
			FileName: "github.com/example/project/cmd/server/main.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 10, Count: 1},
				{StartLine: 11, StartCol: 1, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 0},
			},
		},
		{
			FileName: "github.com/example/project/internal/service/user.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 1, Count: 1},
			},
		},
	}

	// Add more profiles to trigger concurrent processing
	for i := 0; i < 20; i++ {
		profiles = append(profiles, &cover.Profile{
			FileName: "github.com/example/project/pkg/test/file" + string(rune('a'+i)) + ".go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 1, Count: 1},
			},
		})
	}

	tests := []struct {
		name           string
		level          int
		ignorePatterns []string
	}{
		{
			name:  "concurrent with level 0",
			level: 0,
		},
		{
			name:  "concurrent with level -1",
			level: -1,
		},
		{
			name:           "concurrent with ignore patterns",
			level:          0,
			ignorePatterns: []string{"*/internal/*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCoverageAnalyzer(tt.level, tt.ignorePatterns)

			// Get results from both sequential and concurrent methods
			seqResult := analyzer.Aggregate(profiles)
			concResult := analyzer.AggregateConcurrent(profiles)

			// Compare results - they should be identical
			if !reflect.DeepEqual(seqResult, concResult) {
				t.Errorf("Concurrent result differs from sequential result")
				t.Errorf("Sequential: %+v", seqResult)
				t.Errorf("Concurrent: %+v", concResult)
			}
		})
	}
}

func TestAggregateConcurrentSmallInput(t *testing.T) {
	// Test that small inputs fall back to sequential processing
	profiles := []*cover.Profile{
		{
			FileName: "test/file.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
			},
		},
	}

	analyzer := NewCoverageAnalyzer(0, nil)
	result := analyzer.AggregateConcurrent(profiles)

	expected := map[string]*DirCoverage{
		"test": {
			Dir:         "test",
			StmtCount:   5,
			StmtCovered: 5,
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func BenchmarkAggregate(b *testing.B) {
	// Create a large set of test profiles
	var profiles []*cover.Profile
	for i := 0; i < 100; i++ {
		profiles = append(profiles, &cover.Profile{
			FileName: "github.com/example/project/pkg/module" + string(rune('a'+i%26)) + "/file.go",
			Mode:     "set",
			Blocks: []cover.ProfileBlock{
				{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 10, Count: 1},
				{StartLine: 11, StartCol: 1, EndLine: 20, EndCol: 1, NumStmt: 10, Count: 0},
			},
		})
	}

	analyzer := NewCoverageAnalyzer(0, nil)

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = analyzer.Aggregate(profiles)
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = analyzer.AggregateConcurrent(profiles)
		}
	})
}
