package main

import (
	"testing"
)

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   *HunkInfo
	}{
		{
			name:   "simple hunk",
			header: "@@ -1,3 +1,5 @@",
			want: &HunkInfo{
				OldStart: 1,
				OldCount: 3,
				NewStart: 1,
				NewCount: 5,
			},
		},
		{
			name:   "single line hunk",
			header: "@@ -10 +15 @@",
			want: &HunkInfo{
				OldStart: 10,
				OldCount: 1,
				NewStart: 15,
				NewCount: 1,
			},
		},
		{
			name:   "complex hunk",
			header: "@@ -100,20 +150,30 @@ func main() {",
			want: &HunkInfo{
				OldStart: 100,
				OldCount: 20,
				NewStart: 150,
				NewCount: 30,
			},
		},
		{
			name:   "invalid header",
			header: "not a hunk header",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHunkHeader(tt.header)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseHunkHeader() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("parseHunkHeader() = nil, want %v", tt.want)
				return
			}
			if got.OldStart != tt.want.OldStart || got.OldCount != tt.want.OldCount ||
				got.NewStart != tt.want.NewStart || got.NewCount != tt.want.NewCount {
				t.Errorf("parseHunkHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAddedLinesFromHunk(t *testing.T) {
	tests := []struct {
		name       string
		allLines   []string
		hunkHeader string
		info       *HunkInfo
		want       []int
	}{
		{
			name: "simple addition",
			allLines: []string{
				"@@ -1,2 +1,4 @@",
				" line1",
				" line2",
				"+line3",
				"+line4",
			},
			hunkHeader: "@@ -1,2 +1,4 @@",
			info: &HunkInfo{
				OldStart: 1,
				OldCount: 2,
				NewStart: 1,
				NewCount: 4,
			},
			want: []int{3, 4},
		},
		{
			name: "mixed changes",
			allLines: []string{
				"@@ -1,3 +1,3 @@",
				"-old line",
				"+new line",
				" unchanged",
				"+another new",
			},
			hunkHeader: "@@ -1,3 +1,3 @@",
			info: &HunkInfo{
				OldStart: 1,
				OldCount: 3,
				NewStart: 1,
				NewCount: 3,
			},
			want: []int{1, 3},
		},
		{
			name: "no additions",
			allLines: []string{
				"@@ -1,3 +1,2 @@",
				"-removed line",
				" kept line",
				" another kept",
			},
			hunkHeader: "@@ -1,3 +1,2 @@",
			info: &HunkInfo{
				OldStart: 1,
				OldCount: 3,
				NewStart: 1,
				NewCount: 2,
			},
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAddedLinesFromHunk(tt.allLines, tt.hunkHeader, tt.info)
			if len(got) != len(tt.want) {
				t.Errorf("getAddedLinesFromHunk() returned %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("getAddedLinesFromHunk()[%d] = %d, want %d", i, line, tt.want[i])
				}
			}
		})
	}
}

func TestParseFileDiff(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		diffContent string
		want        []DiffLine
	}{
		{
			name:     "simple diff",
			filename: "main.go",
			diffContent: `@@ -1,2 +1,3 @@
 func main() {
+	fmt.Println("Hello")
 }`,
			want: []DiffLine{
				{File: "main.go", LineNum: 2, ChangeType: "added"},
			},
		},
		{
			name:     "multiple hunks",
			filename: "util.go",
			diffContent: `@@ -1,2 +1,3 @@
 package main
+import "fmt"
 
@@ -10,3 +11,5 @@
 func helper() {
+	fmt.Println("helper")
+	return nil
 }`,
			want: []DiffLine{
				{File: "util.go", LineNum: 2, ChangeType: "added"},
				{File: "util.go", LineNum: 12, ChangeType: "added"},
				{File: "util.go", LineNum: 13, ChangeType: "added"},
			},
		},
		{
			name:     "with deletions",
			filename: "test.go",
			diffContent: `@@ -5,4 +5,3 @@
 func test() {
-	oldCode()
+	newCode()
 	done()
 }`,
			want: []DiffLine{
				{File: "test.go", LineNum: 6, ChangeType: "added"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFileDiff(tt.filename, tt.diffContent)
			if len(got) != len(tt.want) {
				t.Errorf("parseFileDiff() returned %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i, line := range got {
				if line.File != tt.want[i].File || line.LineNum != tt.want[i].LineNum ||
					line.ChangeType != tt.want[i].ChangeType {
					t.Errorf("parseFileDiff()[%d] = %v, want %v", i, line, tt.want[i])
				}
			}
		})
	}
}

func TestNormalizeFilePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "relative path",
			path: "main.go",
			want: "main.go",
		},
		{
			name: "relative with dot",
			path: "./main.go",
			want: "main.go",
		},
		{
			name: "absolute path",
			path: "/home/user/project/main.go",
			want: "/home/user/project/main.go",
		},
		{
			name: "nested relative",
			path: "internal/service/user.go",
			want: "internal/service/user.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeFilePath(tt.path)
			if got != tt.want {
				t.Errorf("normalizeFilePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// Mock diff output for testing
const mockDiffOutput = `diff --git a/main.go b/main.go
index abc123..def456 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,5 @@
 package main
 
+import "fmt"
+
 func main() {
@@ -10,2 +12,3 @@ func main() {
 	// existing code
+	fmt.Println("Hello, World!")
 }`

// TestGetGitDiff tests git diff parsing
// Note: This test requires manual mocking or will be skipped in environments without git
func TestGetGitDiff(t *testing.T) {
	// Skip if not in a git repository
	if _, err := getMergeBase(); err != nil {
		t.Skip("Skipping git-dependent test - not in a git repository")
	}

	tests := []struct {
		name    string
		baseRef string
		wantErr bool
	}{
		{
			name:    "with explicit base",
			baseRef: "HEAD",
			wantErr: false,
		},
		{
			name:    "empty base (default)",
			baseRef: "",
			wantErr: false,
		},
		{
			name:    "staged changes",
			baseRef: "staged",
			wantErr: false,
		},
		{
			name:    "working changes",
			baseRef: "working",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGitDiff(tt.baseRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGitDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("GetGitDiff() returned nil without error")
			}
		})
	}
}
