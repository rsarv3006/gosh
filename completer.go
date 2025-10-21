//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/chzyer/readline"
)

// GoshCompleter implements the readline.AutoCompleter interface with intelligent Go capabilities
type GoshCompleter struct {
	contextAnalyzer *ContextAnalyzer
	symbolExtractor *SymbolExtractor
	goEvaluator     *GoEvaluator
	lspWrapper      *LSPClientWrapper
	lspEnabled      bool
}

// NewGoshCompleter creates a new intelligent completer
func NewGoshCompleter(goEvaluator *GoEvaluator) readline.AutoCompleter {
	// Try to initialize LSP wrapper with timeout
	var lspWrapper *LSPClientWrapper
	lspEnabled := false

	// Start LSP in goroutine to avoid blocking startup
	lspChan := make(chan *LSPClientWrapper, 1)
	errChan := make(chan error, 1)

	go func() {
		if lsp, err := NewLSPClientWrapper(); err == nil {
			lspChan <- lsp
		} else {
			errChan <- err
		}
	}()

	// Wait for LSP initialization with timeout
	select {
	case lsp := <-lspChan:
		lspWrapper = lsp
		lspEnabled = true
		fmt.Fprintf(os.Stderr, "✨ LSP intellisense enabled!\n")
	case err := <-errChan:
		// LSP not available, fall back to basic completion
		fmt.Fprintf(os.Stderr, "Note: LSP intellisense unavailable (%v). Using basic Go completion.\n", err)
		lspWrapper = nil
	case <-time.After(5000 * time.Millisecond):
		// Timeout, proceed without LSP
		fmt.Fprintf(os.Stderr, "Note: LSP intellisense starting slowly. Using basic Go completion for now.\n")
		lspWrapper = nil
	}

	return &GoshCompleter{
		contextAnalyzer: NewContextAnalyzer(),
		symbolExtractor: NewSymbolExtractor(goEvaluator.interp),
		goEvaluator:     goEvaluator,
		lspWrapper:      lspWrapper,
		lspEnabled:      lspEnabled,
	}
}

// NewGoshCompleterForTesting creates a new completer for testing (returns concrete type)
func NewGoshCompleterForTesting(goEvaluator *GoEvaluator) *GoshCompleter {
	return &GoshCompleter{
		contextAnalyzer: NewContextAnalyzer(),
		symbolExtractor: NewSymbolExtractor(goEvaluator.interp),
		goEvaluator:     goEvaluator,
		lspWrapper:      nil,   // No LSP for testing
		lspEnabled:      false, // Disabled for testing
	}
}

// GetLSPClient returns the LSP client if available
func (g *GoshCompleter) GetLSPClient() *LSPClientWrapper {
	return g.lspWrapper
}

// Do implements the readline.AutoCompleter interface with intelligent Go completion
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
	lineStr := string(line[:wordStart])
	prefixWords := strings.Fields(lineStr)

	var matches [][]rune

	// Check if we should use intelligent Go completion
	isGo := g.contextAnalyzer.IsGoContext(string(line), pos)
	fmt.Fprintf(os.Stderr, "🔍 [COMPLETER] Line: %q, Pos: %d, IsGo: %v\n", string(line), pos, isGo)

	if isGo {
		// Use intelligent Go completion
		fmt.Fprintf(os.Stderr, "✅ [COMPLETER] Using Go completion for %q\n", partial)
		// Pass the full line, not just the prefix
		fullLine := string(line)
		matches = g.doGoCompletion(fullLine, partial, pos)
	} else if len(prefixWords) == 0 {
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

// doGoCompletion performs intelligent Go code completion with LSP support
func (g *GoshCompleter) doGoCompletion(lineStr, partial string, pos int) [][]rune {
	// Try LSP completion first if available
	if g.lspEnabled && g.lspWrapper.IsReady() {
		fmt.Fprintf(os.Stderr, "🎯 [COMPLETER] LSP ready, trying LSP completion first\n")
		if lspMatches := g.doLSPCompletion(lineStr, partial, pos); len(lspMatches) > 0 {
			fmt.Fprintf(os.Stderr, "✅ [COMPLETER] LSP provided %d matches, using those\n", len(lspMatches))
			return lspMatches
		}
		fmt.Fprintf(os.Stderr, "⚠️  [COMPLETER] LSP returned no matches, falling back to basic completion\n")
		// If LSP fails or returns empty, fall back to basic completion
	} else {
		if g.lspEnabled {
			fmt.Fprintf(os.Stderr, "⚠️  [COMPLETER] LSP enabled but not ready, using basic completion\n")
		} else {
			fmt.Fprintf(os.Stderr, "ℹ️  [COMPLETER] LSP disabled, using basic completion\n")
		}
	}

	// Analyze the context for intelligent completion
	ctx := g.contextAnalyzer.AnalyzeContext(lineStr, pos)

	// Refresh symbol cache if needed
	g.symbolExtractor.refreshIfNeeded()

	var suggestions []CompletionItem

	switch ctx.Type {
	case ContextPackageImport:
		suggestions = g.contextAnalyzer.GetStandardPackages()
	case ContextSelector:
		// Get selector completions (e.g., "fmt.", "strings.")
		suggestions = g.contextAnalyzer.GetSelectorCompletions(ctx.Scope, partial)
		if len(suggestions) == 0 {
			// Fallback to symbol extractor for user-defined symbols
			suggestions = g.symbolExtractor.GetSelectorCompletions(ctx.Scope, partial)
		}
	case ContextVariableDeclaration:
		suggestions = g.symbolExtractor.GetVariables(partial)
		if len(suggestions) == 0 {
			// Fallback to general completions
			suggestions = g.contextAnalyzer.GetVariableCompletions(partial)
		}
	case ContextFunctionCall:
		suggestions = g.symbolExtractor.GetFunctions(partial)
		if len(suggestions) == 0 {
			// Fallback to general function completions
			suggestions = g.contextAnalyzer.GetFunctionCompletions(partial)
		}
	case ContextTypeDeclaration:
		suggestions = g.symbolExtractor.GetTypes(partial)
		if len(suggestions) == 0 {
			// Add built-in types
			suggestions = append(suggestions, CompletionItem{
				Label: "string", Kind: "type", Detail: "String type",
			})
			suggestions = append(suggestions, CompletionItem{
				Label: "int", Kind: "type", Detail: "Integer type",
			})
			suggestions = append(suggestions, CompletionItem{
				Label: "bool", Kind: "type", Detail: "Boolean type",
			})
		}
	default:
		// General Go completion
		suggestions = g.symbolExtractor.GetCompletionSuggestions(partial)
		if len(suggestions) == 0 {
			suggestions = g.contextAnalyzer.GetCompletionSuggestions(ctx)
		}

		// Add Go keywords for common patterns
		if strings.HasPrefix("func", partial) {
			suggestions = append(suggestions, CompletionItem{
				Label:  "func",
				Kind:   "keyword",
				Detail: "function keyword",
			})
		}
		if strings.HasPrefix("return", partial) {
			suggestions = append(suggestions, CompletionItem{
				Label:  "return",
				Kind:   "keyword",
				Detail: "return keyword",
			})
		}
		if strings.HasPrefix("var", partial) {
			suggestions = append(suggestions, CompletionItem{
				Label:  "var",
				Kind:   "keyword",
				Detail: "variable declaration",
			})
		}
		if strings.HasPrefix("if", partial) {
			suggestions = append(suggestions, CompletionItem{
				Label:  "if",
				Kind:   "keyword",
				Detail: "conditional statement",
			})
		}
	}

	// Convert suggestions to rune slices for readline - accept gopls results as-is
	var matches [][]rune
	fmt.Fprintf(os.Stderr, "💡 [COMPLETER] Got %d suggestions from gopls for partial %q\n", len(suggestions), partial)

	for _, suggestion := range suggestions {
		// Accept all gopls results - let gopls handle the filtering
		fmt.Fprintf(os.Stderr, "  ✅ [COMPLETER] Accepting gopls result: %q (kind: %s)\n", suggestion.Label, suggestion.Kind)

		// For prefix matching, calculate suffix
		var suffix string
		if strings.HasPrefix(suggestion.Label, partial) {
			suffix = suggestion.Label[len(partial):]
		} else {
			// For non-prefix matches, replace the entire input
			suffix = suggestion.Label
		}
		matches = append(matches, []rune(suffix))
	}

	fmt.Fprintf(os.Stderr, "📤 [COMPLETER] Returning %d matches to readline\n", len(matches))

	return matches
}

// doLSPCompletion performs LSP-based completion
func (g *GoshCompleter) doLSPCompletion(lineStr, partial string, pos int) [][]rune {
	fmt.Fprintf(os.Stderr, "🚀 [COMPLETER] Trying LSP-based completion for: %q (partial: %q)\n", lineStr, partial)

	// Get completions from gopls
	lspItems, err := g.lspWrapper.GetCompletions(lineStr, pos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ [COMPLETER] LSP completion failed: %v - falling back to basic completion\n", err)
		return nil // LSP failed, fall back to basic completion
	}

	fmt.Fprintf(os.Stderr, "✅ [COMPLETER] LSP returned %d items for %q\n", len(lspItems), partial)

	// Convert to our format
	suggestions := ConvertLSPCompletions(lspItems)

	// Filter and convert to rune slices for readline
	var matches [][]rune
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion.Label, partial) {
			suffix := suggestion.Label[len(partial):]
			matches = append(matches, []rune(suffix))
			fmt.Fprintf(os.Stderr, "  ➡️  [COMPLETER] LSP match: %q -> suffix: %q\n", suggestion.Label, suffix)
		}
	}

	fmt.Fprintf(os.Stderr, "📤 [COMPLETER] LSP returning %d matches for partial %q\n", len(matches), partial)
	return matches
}

// cleanup shuts down the LSP client if it was initialized
func (g *GoshCompleter) cleanup() {
	if g.lspEnabled && g.lspWrapper != nil {
		if err := g.lspWrapper.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to shutdown LSP client: %v\n", err)
		}
	}
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
