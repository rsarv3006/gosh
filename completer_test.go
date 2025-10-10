//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
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
			pos:           6,
			expectedMatch: "", // Multiple matches: cd, command
		},
		{
			name:          "Complete help topic 'g'",
			line:          "help g",
			pos:           6,
			expectedMatch: "", // Multiple matches: go, golang
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
			} else {
				// When expectedMatch is empty, we expect multiple matches
				if len(matches) == 0 {
					t.Errorf("expected multiple matches, got %d", len(matches))
				}
			}
			
			// Verify position is within bounds to prevent panic
			if tt.pos > len(tt.line) {
				t.Errorf("test error: pos %d is longer than line length %d", tt.pos, len(tt.line))
				return
			}
			
			// Verify length matches the partial word being completed
			// Find the word being completed (from last space to cursor)
			lastSpace := strings.LastIndex(tt.line[:tt.pos], " ")
			if lastSpace == -1 {
				lastSpace = 0
			} else {
				lastSpace++ // Skip the space
			}
			partialWord := tt.line[lastSpace:tt.pos]
			expectedLength := len(partialWord)
			if length != expectedLength {
				t.Errorf("expected length %d (for partial word '%s'), got %d", expectedLength, partialWord, length)
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
		expectMatches bool  // true if we expect some matches, false for no matches
	}{
		{
			name:          "Complete 'f' in current directory",
			line:          "ls f",
			pos:           4,
			expectMatches: true, // Should find file1.txt, file2.go
		},
		{
			name:          "Complete 'file1' in current directory",
			line:          "cat file1",
			pos:           9,
			expectMatches: true, // Should find file1.txt
		},
		{
			name:          "Complete 'dir' to 'directory/'",
			line:          "cd dir",
			pos:           6,
			expectMatches: true, // Should find directory/
		},
		{
			name:          "Complete with no matches",
			line:          "ls xyz",
			pos:           6,
			expectMatches: false, // Should find no matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, length := c.Do([]rune(tt.line), tt.pos)
			
			if tt.expectMatches {
				if len(matches) == 0 {
					t.Errorf("expected some matches, got %d", len(matches))
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("expected no matches, got %d", len(matches))
				}
			}
			
			// Verify position is within bounds to prevent panic
			if tt.pos > len(tt.line) {
				t.Errorf("test error: pos %d is longer than line length %d", tt.pos, len(tt.line))
				return
			}
			
			// Verify length matches the partial word being completed
			// Find the word being completed (from last space to cursor)
			lastSpace := strings.LastIndex(tt.line[:tt.pos], " ")
			if lastSpace == -1 {
				lastSpace = 0
			} else {
				lastSpace++ // Skip the space
			}
			partialWord := tt.line[lastSpace:tt.pos]
			expectedLength := len(partialWord)
			if length != expectedLength {
				t.Errorf("expected length %d (for partial word '%s'), got %d", expectedLength, partialWord, length)
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
