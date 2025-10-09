//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"

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
	lineStr := string(line[:pos])
	
	// Split the line into words to find the current word
	words := strings.Fields(lineStr)
	var lastWord string
	var wordStart int
	
	if len(words) == 0 {
		// No words yet, suggest commands
		return g.completeCommands("", 0)
	}
	
	// Find the current word being completed
	if strings.HasSuffix(lineStr, " ") {
		// Completing a new word after a space
		lastWord = ""
		wordStart = pos
	} else {
		// Completing the last word
		lastWord = words[len(words)-1]
		wordStart = strings.LastIndex(lineStr[:pos], lastWord)
	}
	
	// Determine completion context
	if len(words) == 1 {
		// First word - suggest commands
		return g.completeCommands(lastWord, wordStart)
	} else {
		// Arguments - suggest files based on command
		cmd := words[0]
		return g.completeArguments(cmd, lastWord, wordStart)
	}
}

// completeCommands provides command completion
func (g *GoshCompleter) completeCommands(partial string, start int) ([][]rune, int) {
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

	return matches, start
}

// completeArguments provides argument completion
func (g *GoshCompleter) completeArguments(cmd, partial string, start int) ([][]rune, int) {
	if cmd == "cd" {
		return g.completeFiles(partial, start, true) // Directories only
	}
	
	// For commands that take files
	if cmd == "ls" || cmd == "cat" || cmd == "head" || cmd == "tail" || cmd == "grep" {
		return g.completeFiles(partial, start, false) // All files
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
		return matches, start
	}
	
	// Default to file completion for unknown commands
	return g.completeFiles(partial, start, false)
}

// completeFiles provides file/directory completion
func (g *GoshCompleter) completeFiles(partial string, start int, dirsOnly bool) ([][]rune, int) {
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
		return matches, start
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

	return matches, start
}
