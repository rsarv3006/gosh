//go:build darwin || linux

package main

import (
	"os"
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

// NewGoshCompleterForTesting creates a new completer for testing (returns concrete type)
func NewGoshCompleterForTesting() *GoshCompleter {
	return &GoshCompleter{}
}

// Do implements the readline.AutoCompleteCompleter interface
func (g *GoshCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	wordStart := pos
	for wordStart > 0 && wordStart <= len(line) {
		if unicode.IsSpace(line[wordStart-1]) {
			break
		}
		wordStart--
	}

	partialRunes := line[wordStart:pos]
	partial := string(partialRunes)

	prefixWords := strings.Fields(string(line[:wordStart]))

	var matches [][]rune
	
	if len(prefixWords) == 0 {
		// Command completion - strip ./ prefix if present
		commandPartial := partial
		if strings.HasPrefix(partial, "./") {
			commandPartial = partial[2:] // Remove "./" prefix
		}
		matches = g.completeCommands(commandPartial)
	} else {
		cmd := prefixWords[0]
		matches = g.completeArguments(cmd, partial)
	}
	
	// For readline AutoCompleter, we need to return the completions as-is.
	// The library will handle the replacement logic correctly.
	return matches, len(partialRunes)
}

// completeCommands provides command completion
func (g *GoshCompleter) completeCommands(partial string) [][]rune {
	var matches [][]rune

	if partial == "" {
		return matches // Return empty matches list
	}

	// 1. Builtin commands
	builtins := []string{"cd", "pwd", "exit", "help"}
	for _, cmd := range builtins {
		if strings.HasPrefix(cmd, partial) {
			suffix := cmd[len(partial):]
			matches = append(matches, []rune(suffix))
		}
	}

	// 2. Commands from PATH
	if path, ok := os.LookupEnv("PATH"); ok {
		pathCommands := g.getCommandsFromPath(path, partial)
		matches = append(matches, pathCommands...)
	}

	// 3. Local directory executables (including this is key!)
	localCommands := g.getLocalExecutables(partial)
	matches = append(matches, localCommands...)

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
				// Return only the suffix that needs to be added
				suffix := topic[len(partial):]
				matches = append(matches, []rune(suffix))
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

	// Handle ~ expansion separately - detect early to avoid path parsing conflicts
	var isTildePath bool
	var homeDir string
	originalPartialForSuffix := partial
	
	if strings.HasPrefix(partial, "~") {
		isTildePath = true
		homeDir = os.Getenv("HOME")
	}

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
	}

	// Handle ~ expansion for directory lookup
	if isTildePath {
		if strings.HasPrefix(dir, "~") {
			dir = strings.Replace(dir, "~", homeDir, 1)
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
			completionName := name
			if file.IsDir() {
				completionName += "/"
			}

			// Calculate the suffix to return
			var suffix string
			if lastSlash == -1 {
				// Simple filename completion - return suffix of the filename
				suffix = completionName[len(pattern):]
			} else {
				// Path completion - reconstruct the path and calculate suffix
				var completedPath string
				if isTildePath {
					// Need to convert back to ~ format for the user
					userPath := strings.Replace(dir, homeDir, "~", 1)
					completedPath = userPath + "/" + completionName
				} else {
					completedPath = dir + "/" + completionName	
				}
				suffix = completedPath[len(originalPartialForSuffix):]
			}

			matches = append(matches, []rune(suffix))
		}
	}

	return matches
}

// getCommandsFromPath finds executables in PATH directories that match partial
func (g *GoshCompleter) getCommandsFromPath(pathEnv, partial string) [][]rune {
	var matches [][]rune
	
	// Split PATH by colon
	pathDirs := strings.Split(pathEnv, ":")
	seen := make(map[string]bool) // Avoid duplicates

	for _, dir := range pathDirs {
		if dir == "" {
			continue
		}
		
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			name := entry.Name()
			
			// Check if it's executable (Unix check)
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			if info.Mode().Perm()&0111 == 0 { // Not executable
				continue
			}
			
			// Skip if already seen
			if seen[name] {
				continue
			}
			seen[name] = true
			
			if strings.HasPrefix(name, partial) {
				suffix := name[len(partial):]
				matches = append(matches, []rune(suffix))
			}
		}
	}
	
	return matches
}

// getLocalExecutables finds executables in current directory
func (g *GoshCompleter) getLocalExecutables(partial string) [][]rune {
	var matches [][]rune
	
	entries, err := os.ReadDir(".")
	if err != nil {
		return matches
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		
		// Check if it's executable
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if info.Mode().Perm()&0111 == 0 { // Not executable
			continue
		}
		
		if strings.HasPrefix(name, partial) {
			suffix := name[len(partial):]
			matches = append(matches, []rune(suffix))
		}
	}
	
	return matches
}
