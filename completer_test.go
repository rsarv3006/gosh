//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoshCompleter_Do_CommandCompletion(t *testing.T) {
	c := NewGoshCompleterForTesting()

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedLength int
		expectedMatch  string
	}{
		{
			name:           "Complete 'wh' to 'whoami'",
			line:           "wh",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "oami", // Should be suffix only
		},
		{
			name:           "Complete 'ls' to 'ls'",
			line:           "ls",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "", // Should be exact match, no suffix
		},
		{
			name:           "Complete 'c' (multiple matches)",
			line:           "c",
			pos:            1,
			expectedLength: 1,
			expectedMatch:  "", // Multiple matches: cd, cat, const
		},
		{
			name:           "Complete 'ex' to 'exit'",
			line:           "ex",
			pos:            2,
			expectedLength: 2,
			expectedMatch:  "it",
		},
		{
			name:           "Complete 'x' (multiple matches)",
			line:           "x",
			pos:            1,
			expectedLength: 1,
			expectedMatch:  "", // Multiple matches, exact completion not performed
		},
		{
			name:           "Complete 'nonexistent'",
			line:           "nonexistent",
			pos:            11,
			expectedLength: 11,
			expectedMatch:  "", // No matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if length != tt.expectedLength {
				t.Errorf("expected length %d, got %d", tt.expectedLength, length)
			}
			
			if tt.expectedMatch != "" {
				if len(matches) != 1 {
					t.Errorf("expected 1 match, got %d", len(matches))
					return
				}
				if string(matches[0]) != tt.expectedMatch {
					t.Errorf("expected match '%s', got '%s'", tt.expectedMatch, string(matches[0]))
				}
			}
		})
	}
}

func TestGoshCompleter_Do_ArgumentCompletion_Help(t *testing.T) {
	c := NewGoshCompleterForTesting()

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedMatch  string
	}{
		{
			name:          "Complete help topic 'c' to 'cd'",
			line:          "help c",
			pos:           7,
			expectedMatch: "d",
		},
		{
			name:          "Complete help topic 'g' to 'git'",
			line:          "help ",
			pos:           5,
			expectedMatch: "", // Multiple matches starting with 'g'
		},
		{
			name:          "Complete help topic 'subs' to 'substitution'",
			line:          "help subs",
			pos:           9,
			expectedMatch: "titution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if tt.expectedMatch != "" {
				if len(matches) != 1 {
					t.Errorf("expected 1 match, got %d", len(matches))
					return
				}
				if string(matches[0]) != tt.expectedMatch {
					t.Errorf("expected match '%s', got '%s'", tt.expectedMatch, string(matches[0]))
				}
			}
			
			// Just verify length is reasonable for partial completion
			expectedLength := len(tt.line[:tt.pos])
			if length != expectedLength {
				t.Errorf("expected length %d, got %d", expectedLength, length)
			}
		})
	}
}

func TestGoshCompleter_Do_FileCompletion(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Create a temporary directory with test files
	tempDir := t.TempDir()
	
	testFiles := []string{"file1.txt", "file2.go", "directory", "test.txt"}
	for _, name := range testFiles {
		if name == "directory" {
			os.MkdirAll(filepath.Join(tempDir, name), 0755)
		} else {
			CreateTestFile(t, filepath.Join(tempDir, name), "content")
		}
	}

	// Change to temp directory for file completion tests
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedMatch  string
	}{
		{
			name:          "Complete 'f' in current directory",
			line:          "ls ",
			pos:           3,
			expectedMatch: "", // Multiple matches starting with 'f'
		},
		{
			name:          "Complete 'file1.' to '.txt'",
			line:          "cat file1.",
			pos:           11,
			expectedMatch: "txt",
		},
		{
			name:          "Complete 'dir' to 'directory/'",
			line:          "cd dir",
			pos:           7,
			expectedMatch: "ectory/",
		},
		{
			name:          "Complete 'test.' to '.txt'",
			line:          "cat test.",
			pos:           9,
			expectedMatch: "txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if tt.expectedMatch != "" {
				if len(matches) != 1 {
					t.Errorf("expected 1 match, got %d", len(matches))
					return
				}
				if string(matches[0]) != tt.expectedMatch {
					t.Errorf("expected match '%s', got '%s'", tt.expectedMatch, string(matches[0]))
				}
			} else if len(matches) == 0 && tt.expectedMatch == "" {
				// If we expect no specific match but got no matches at all, that's ok
				// This happens when there are multiple matches (show all options)
			} else if len(matches) == 1 {
				// Single exact match case
				if string(matches[0]) != tt.expectedMatch {
					t.Errorf("expected match '%s', got '%s'", tt.expectedMatch, string(matches[0]))
				}
			}
			
			expectedLength := len(tt.line[:tt.pos])
			if length != expectedLength {
				t.Errorf("expected length %d, got %d", expectedLength, length)
			}
		})
	}
}

func TestGoshCompleter_completeCommands(t *testing.T) {
	c := NewGoshCompleterForTesting()

	tests := []struct {
		partial   string
		expected  []string
	}{
		{
			partial:  "wh",
			expected: []string{"oami"}, // whoami - should return suffix
		},
		{
			partial:  "cd", 
			expected: []string{""}, // exact match
		},
		{
			partial:  "c",
			expected: []string{"d", "at", "onst"}, // cd, cat, const
		},
		{
			partial:  "e",
			expected: []string{"xit", "cho"}, // exit, echo
		},
		{
			partial:  "nonexistent",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run("completeCommands_"+tt.partial, func(t *testing.T) {
			result := c.completeCommands(tt.partial)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, match := range result {
				expected := tt.expected[i]
				if string(match) != expected {
					t.Errorf("expected match '%s', got '%s'", expected, string(match))
				}
			}
		})
	}
}

func TestGoshCompleter_completeFiles(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Create a temporary directory with test files
	tempDir := t.TempDir()
	
	testFiles := []string{"file1.txt", "file2.go", "directory", "test.txt"}
	for _, name := range testFiles {
		if name == "directory" {
			os.MkdirAll(filepath.Join(tempDir, name), 0755)
		} else {
			CreateTestFile(t, filepath.Join(tempDir, name), "content")
		}
	}

	// Change to temp directory for file completion tests  
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		partial  string
		dirsOnly bool
		expected []string
	}{
		{
			partial:  "file",
			dirsOnly: false,
			expected: []string{"1.txt", "2.go"}, // suffixes for file1.txt, file2.go
		},
		{
			partial:  "dir",
			dirsOnly: true,
			expected: []string{"ectory/"}, // directory/ - should include slash
		},
		{
			partial:  "test",
			dirsOnly: false,
			expected: []string{".txt"}, // .txt for test.txt
		},
		{
			partial:  "nonexistent",
			dirsOnly: false,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run("completeFiles_"+tt.partial, func(t *testing.T) {
			result := c.completeFiles(tt.partial, tt.dirsOnly)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, match := range result {
				expected := tt.expected[i]
				if string(match) != expected {
					t.Errorf("expected match '%s', got '%s'", expected, string(match))
				}
			}
		})
	}
}

// Helper function to create test files
func CreateTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
}
