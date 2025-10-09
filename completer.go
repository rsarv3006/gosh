//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/chzyer/readline"
)

// GoshCompleter implements the readline.AutoCompleteCompleter interface
type GoshCompleter struct{}

// NewGoshCompleter creates a new completer
func NewGoshCompleter() readline.AutoCompleter {
	return &GoshCompleter{}
}

// Do implements the readline.AutoCompleteCompleter interface
func (g *GoshCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	wordStart := pos
	for wordStart > 0 {
		if unicode.IsSpace(line[wordStart-1]) {
			break
		}
		wordStart--
	}

	partialRunes := line[wordStart:pos]
	partial := string(partialRunes)

	prefixWords := strings.Fields(string(line[:wordStart]))

	if len(prefixWords) == 0 {
		matches := g.completeCommands(partial)
		return matches, len(partialRunes)
	}

	cmd := prefixWords[0]
	matches := g.completeArguments(cmd, partial)
	return matches, len(partialRunes)
}

// completeCommands provides command completion
func (g *GoshCompleter) completeCommands(partial string) [][]rune {
	var matches [][]rune

	// Commands to complete
	commands := []string{
		"cd", "pwd", "exit", "help",
		"ls", "cat", "grep", "find", "ps", "kill",
		"echo", "date", "whoami", "head", "tail",
		"func", "var", "const", "type", "import",
		"for", "if", "switch", "select", "return", "go", "defer",
	}

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, partial) {
			matches = append(matches, []rune(cmd))
		}
	}

	return matches
}

// completeArguments provides argument completion
func (g *GoshCompleter) completeArguments(cmd, partial string) [][]rune {
	if cmd == "cd" {
		return g.completeFiles(partial, true) // Directories only
	}

	// For commands that take files
	if cmd == "ls" || cmd == "cat" || cmd == "head" || cmd == "tail" || cmd == "grep" {
		return g.completeFiles(partial, false) // All files
	}

	// For help command
	if cmd == "help" {
		topics := []string{"cd", "pwd", "exit", "help", "go", "golang", "yaegi", "substitution", "command"}
		var matches [][]rune
		for _, topic := range topics {
			if strings.HasPrefix(topic, partial) {
				matches = append(matches, []rune(topic))
			}
		}
		return matches
	}

	// Default to file completion for unknown commands
	return g.completeFiles(partial, false)
}

// completeFiles provides file/directory completion
func (g *GoshCompleter) completeFiles(partial string, dirsOnly bool) [][]rune {
	var matches [][]rune

	// Extract directory and file pattern
	dir := "."
	pattern := partial
	var lastSlash int

	if lastSlash = strings.LastIndex(partial, "/"); lastSlash != -1 {
		dir = partial[:lastSlash]
		if dir == "" {
			dir = "/"
		}
		pattern = partial[lastSlash+1:]
	} else if strings.HasPrefix(partial, "~") {
		// Handle home directory
		dir = os.Getenv("HOME")
		if len(partial) > 1 {
			remaining := partial[1:]
			if lastSlash := strings.LastIndex(remaining, "/"); lastSlash != -1 {
				dir = filepath.Join(dir, remaining[:lastSlash])
				pattern = remaining[lastSlash+1:]
			} else {
				pattern = remaining
			}
		}
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return matches
	}

	for _, file := range files {
		if dirsOnly && !file.IsDir() {
			continue
		}

		name := file.Name()
		if strings.HasPrefix(name, pattern) {
			// Add trailing slash for directories
			if file.IsDir() {
				name += "/"
			}

			// Reconstruct full path
			var fullName string
			if lastSlash == -1 {
				fullName = name
			} else {
				fullName = partial[:lastSlash+1] + name
			}

			matches = append(matches, []rune(fullName))
		}
	}

	return matches
}
