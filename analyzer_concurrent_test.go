package main

import (
	"fmt"
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

func TestAggregateConcurrentErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		profiles []*cover.Profile
		wantErr  bool
	}{
		{
			name: "nil profile in list",
			profiles: []*cover.Profile{
				{
					FileName: "test/file1.go",
					Mode:     "set",
					Blocks: []cover.ProfileBlock{
						{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
					},
				},
				nil, // This could cause issues if not handled properly
				{
					FileName: "test/file2.go",
					Mode:     "set",
					Blocks: []cover.ProfileBlock{
						{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
					},
				},
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "profile with empty filename",
			profiles: []*cover.Profile{
				{
					FileName: "",
					Mode:     "set",
					Blocks: []cover.ProfileBlock{
						{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
					},
				},
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "profile with nil blocks",
			profiles: []*cover.Profile{
				{
					FileName: "test/file.go",
					Mode:     "set",
					Blocks:   nil,
				},
			},
			wantErr: false, // Should handle gracefully
		},
		{
			name: "large number of profiles to stress test concurrent processing",
			profiles: func() []*cover.Profile {
				var ps []*cover.Profile
				for i := 0; i < 100; i++ {
					ps = append(ps, &cover.Profile{
						FileName: fmt.Sprintf("test/file%d.go", i),
						Mode:     "set",
						Blocks: []cover.ProfileBlock{
							{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
						},
					})
				}
				// Add some problematic profiles in the middle
				ps[50] = nil
				ps[75] = &cover.Profile{FileName: "", Mode: "set"}
				return ps
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCoverageAnalyzer(0, nil)

			// Create enough profiles to trigger concurrent processing
			profiles := tt.profiles
			if len(profiles) < 11 {
				// Add dummy profiles to ensure concurrent processing
				for i := len(profiles); i < 11; i++ {
					profiles = append(profiles, &cover.Profile{
						FileName: fmt.Sprintf("dummy/file%d.go", i),
						Mode:     "set",
						Blocks: []cover.ProfileBlock{
							{StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 1, NumStmt: 1, Count: 1},
						},
					})
				}
			}

			// Run the concurrent aggregation and check it doesn't panic
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("AggregateConcurrent() panicked unexpectedly: %v", r)
				}
			}()

			result := analyzer.AggregateConcurrent(profiles)

			// Verify result is not nil
			if result == nil {
				t.Error("AggregateConcurrent() returned nil result")
			}
		})
	}
}

func TestProcessProfileErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		profile *cover.Profile
		wantErr bool
	}{
		{
			name:    "nil profile",
			profile: nil,
			wantErr: true,
		},
		{
			name: "profile with empty filename",
			profile: &cover.Profile{
				FileName: "",
				Mode:     "set",
				Blocks: []cover.ProfileBlock{
					{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "profile with special characters in filename",
			profile: &cover.Profile{
				FileName: "test/file with spaces and @#$.go",
				Mode:     "set",
				Blocks: []cover.ProfileBlock{
					{StartLine: 1, StartCol: 1, EndLine: 10, EndCol: 1, NumStmt: 5, Count: 1},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCoverageAnalyzer(0, nil)

			// Test processProfile directly
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("processProfile() panicked unexpectedly: %v", r)
					}
				}
			}()

			if tt.profile != nil {
				result := analyzer.processProfile(tt.profile)
				if result == nil {
					t.Error("processProfile() returned nil result")
				}
			}
		})
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
