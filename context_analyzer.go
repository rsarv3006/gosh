//go:build darwin || linux

package main

import (
	"strings"
	"unicode"
)

// ContextType represents the type of completion context
type ContextType int

const (
	ContextUnknown ContextType = iota
	ContextPackageImport
	ContextSelector
	ContextVariableDeclaration
	ContextFunctionCall
	ContextTypeDeclaration
	ContextStructLiteral
	ContextGeneral
)

// CompletionContext represents the context for completion
type CompletionContext struct {
	Type   ContextType
	Scope  string // Package name for selectors, etc.
	Prefix string
	Line   string
	Pos    int
}

// ContextAnalyzer analyzes code to determine completion context
type ContextAnalyzer struct{}

// NewContextAnalyzer creates a new context analyzer
func NewContextAnalyzer() *ContextAnalyzer {
	return &ContextAnalyzer{}
}

// IsGoContext determines if the current line appears to be Go code
func (c *ContextAnalyzer) IsGoContext(line string, pos int) bool {
	if len(line) == 0 {
		return false
	}
	linePrefix := strings.TrimSpace(line[:pos])
	if linePrefix == "" {
		return false
	}

	// Check for shell path patterns FIRST (before other checks)
	if strings.HasPrefix(linePrefix, "./") ||
		strings.HasPrefix(linePrefix, "../") ||
		strings.HasPrefix(linePrefix, "/") ||
		strings.HasPrefix(linePrefix, "~/") {
		return false
	}

	// If it's obviously a shell command (first word is a known command)
	words := strings.Fields(linePrefix)
	if len(words) > 0 {
		shellCommands := []string{"cd", "ls", "pwd", "cat", "grep", "find", "mv", "cp", "rm", "mkdir", "git", "docker", "ps", "kill", "man"}
		for _, cmd := range shellCommands {
			if words[0] == cmd {
				return false
			}
		}
	}

	// Check for obvious Go patterns first
	goPatterns := []string{
		"func ", "var ", "const ", "type ",
		"import ", "package ",
		"if ", "for ", "switch ", "select ",
		"return ", "go ", "defer ",
		"{", "}", "(", ")", ";",
		":=", // Short variable declaration
		"==", "!=", "<", ">", "<=", ">=",
	}
	for _, pattern := range goPatterns {
		if strings.Contains(linePrefix, pattern) {
			return true
		}
	}

	// Check for Go package selector (but not shell paths)
	// Only consider it a package selector if preceded by alphanumeric
	if strings.Contains(linePrefix, ".") {
		for i, ch := range linePrefix {
			if ch == '.' && i > 0 {
				prevChar := linePrefix[i-1]
				// It's a package selector if previous char is alphanumeric or )
				if (prevChar >= 'a' && prevChar <= 'z') ||
					(prevChar >= 'A' && prevChar <= 'Z') ||
					(prevChar >= '0' && prevChar <= '9') ||
					prevChar == ')' || prevChar == ']' {
					return true
				}
			}
		}
	}

	// Check for Go types or patterns
	if strings.Contains(linePrefix, "string") ||
		strings.Contains(linePrefix, "int") ||
		strings.Contains(linePrefix, "bool") ||
		strings.Contains(linePrefix, "[]") ||
		strings.Contains(linePrefix, "map[") ||
		strings.Contains(linePrefix, "chan ") {
		return true
	}

	// Check for comments
	if strings.HasPrefix(linePrefix, "/*") || strings.HasPrefix(linePrefix, "//") {
		return true
	}

	// If it looks like variable assignment (contains = but not obvious shell)
	if strings.Contains(linePrefix, "=") && !strings.Contains(linePrefix, "==") {
		// Check if it looks like Go variable assignment
		// This is a heuristic - if it has camelCase or underscores, probably Go
		for _, word := range words {
			if strings.Contains(word, "_") ||
				(len(word) > 1 && word[0] >= 'a' && word[0] <= 'z') {
				return true
			}
		}
	}

	// Check if it contains common Go function names
	goFunctions := []string{"fmt.", "strings.", "os.", "math.", "time.", "regexp."}
	for _, fn := range goFunctions {
		if strings.Contains(linePrefix, fn) {
			return true
		}
	}

	// Default: if it's not obviously shell and has some complexity, treat as Go
	return len(linePrefix) > 2
}

// AnalyzeContext analyzes the current line and position to determine completion context
func (c *ContextAnalyzer) AnalyzeContext(line string, pos int) CompletionContext {
	if pos > len(line) {
		pos = len(line)
	}

	linePrefix := line[:pos]

	// Check for import context
	if c.isImportContext(linePrefix) {
		return CompletionContext{
			Type:   ContextPackageImport,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Check for selector context (package.member)
	if selectorScope := c.getSelectorScope(linePrefix); selectorScope != "" {
		return CompletionContext{
			Type:   ContextSelector,
			Scope:  selectorScope,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Check for function call context
	if c.isFunctionCallContext(linePrefix) {
		return CompletionContext{
			Type:   ContextFunctionCall,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Check for variable declaration context
	if c.isVariableDeclarationContext(linePrefix) {
		return CompletionContext{
			Type:   ContextVariableDeclaration,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Check for type declaration context
	if c.isTypeDeclarationContext(linePrefix) {
		return CompletionContext{
			Type:   ContextTypeDeclaration,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Check for struct literal context
	if c.isStructLiteralContext(linePrefix) {
		return CompletionContext{
			Type:   ContextStructLiteral,
			Prefix: c.extractPartialWord(linePrefix),
			Line:   line,
			Pos:    pos,
		}
	}

	// Default to general context
	return CompletionContext{
		Type:   ContextGeneral,
		Prefix: c.extractPartialWord(linePrefix),
		Line:   line,
		Pos:    pos,
	}
}

// extractPartialWord extracts the word being completed
func (c *ContextAnalyzer) extractPartialWord(linePrefix string) string {
	start := len(linePrefix)
	for start > 0 {
		r := rune(linePrefix[start-1])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		start--
	}
	return linePrefix[start:]
}

// isImportContext checks if we're in an import statement
func (c *ContextAnalyzer) isImportContext(linePrefix string) bool {
	return strings.Contains(linePrefix, "import ") &&
		!strings.Contains(linePrefix, "\"") &&
		!strings.Contains(linePrefix, ")")
}

// getSelectorScope extracts the package name from a selector expression
func (c *ContextAnalyzer) getSelectorScope(linePrefix string) string {
	lastDot := strings.LastIndex(linePrefix, ".")
	if lastDot == -1 || lastDot == 0 {
		return ""
	}

	// Check if there's a valid identifier before the dot
	scopeStart := lastDot - 1
	scopeEnd := lastDot - 1

	// Find the start of the scope identifier
	for scopeStart >= 0 {
		r := rune(linePrefix[scopeStart])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			scopeStart++
			break
		}
		scopeStart--
	}
	if scopeStart < 0 {
		scopeStart = 0
	}

	// Extract the scope
	if scopeStart <= scopeEnd {
		scope := linePrefix[scopeStart : scopeEnd+1]
		// Check if it looks like a valid identifier
		if c.isValidIdentifier(scope) {
			return scope
		}
	}

	return ""
}

// isValidIdentifier checks if a string is a valid Go identifier
func (c *ContextAnalyzer) isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}

	return true
}

// isFunctionCallContext checks if we're in a function call
func (c *ContextAnalyzer) isFunctionCallContext(linePrefix string) bool {
	return strings.HasSuffix(linePrefix, "(") ||
		strings.HasSuffix(linePrefix, ", ")
}

// isVariableDeclarationContext checks if we're declaring variables
func (c *ContextAnalyzer) isVariableDeclarationContext(linePrefix string) bool {
	return strings.Contains(linePrefix, ":=") ||
		strings.Contains(linePrefix, "var ") ||
		strings.Contains(linePrefix, "const ")
}

// isTypeDeclarationContext checks if we're declaring types
func (c *ContextAnalyzer) isTypeDeclarationContext(linePrefix string) bool {
	return strings.Contains(linePrefix, "type ") &&
		!strings.Contains(linePrefix, "=") &&
		!strings.Contains(linePrefix, "{")
}

// isStructLiteralContext checks if we're in a struct literal
func (c *ContextAnalyzer) isStructLiteralContext(linePrefix string) bool {
	return strings.HasSuffix(linePrefix, "{") ||
		(strings.HasSuffix(linePrefix, ", ") && c.containsStructLiteralStart(linePrefix))
}

// containsStructLiteralStart checks if the line contains the start of a struct literal
func (c *ContextAnalyzer) containsStructLiteralStart(linePrefix string) bool {
	// This is a simplified check - in a more sophisticated implementation,
	// we would use proper AST parsing
	for i := len(linePrefix) - 1; i >= 0; i-- {
		if linePrefix[i] == '{' {
			return true
		}
		if linePrefix[i] == '(' && i > 0 && linePrefix[i-1] == ')' {
			// Type conversion like Type{}
			return true
		}
	}
	return false
}

// GetStandardPackages returns common standard library packages
func (c *ContextAnalyzer) GetStandardPackages() []CompletionItem {
	packages := []CompletionItem{
		{Label: "fmt", Kind: "package", Detail: "Formatting functions", Documentation: "Package fmt implements formatted I/O functions."},
		{Label: "os", Kind: "package", Detail: "Operating system interface", Documentation: "Package os provides a platform-independent interface to operating system functionality."},
		{Label: "strings", Kind: "package", Detail: "String manipulation", Documentation: "Package strings implements simple functions to manipulate strings."},
		{Label: "strconv", Kind: "package", Detail: "String conversion", Documentation: "Package strconv implements conversions to and from string representations."},
		{Label: "path/filepath", Kind: "package", Detail: "File path manipulation", Documentation: "Package filepath implements utility routines for manipulating filename paths."},
		{Label: "encoding/json", Kind: "package", Detail: "JSON encoding/decoding", Documentation: "Package json implements encoding and decoding of JSON."},
		{Label: "net/http", Kind: "package", Detail: "HTTP client/server", Documentation: "Package http provides HTTP client and server implementations."},
		{Label: "time", Kind: "package", Detail: "Time functionality", Documentation: "Package time provides functionality for measuring and displaying time."},
		{Label: "math", Kind: "package", Detail: "Math functions", Documentation: "Package math provides basic constants and mathematical functions."},
		{Label: "regexp", Kind: "package", Detail: "Regular expressions", Documentation: "Package regexp implements regular expression search."},
		{Label: "io", Kind: "package", Detail: "I/O primitives", Documentation: "Package io provides basic interfaces to I/O primitives."},
		{Label: "bufio", Kind: "package", Detail: "Buffered I/O", Documentation: "Package bufio implements buffered I/O."},
		{Label: "log", Kind: "package", Detail: "Logging", Documentation: "Package log implements simple logging."},
		{Label: "container/list", Kind: "package", Detail: "Linked list", Documentation: "Package list implements a doubly linked list."},
		{Label: "container/vector", Kind: "package", Detail: "Vector", Documentation: "Package vector implements vector."},
		{Label: "sort", Kind: "package", Detail: "Sorting", Documentation: "Package sort provides primitives for sorting slices and user-defined collections."},
		{Label: "sync", Kind: "package", Detail: "Synchronization", Documentation: "Package sync provides basic synchronization primitives."},
		{Label: "context", Kind: "package", Detail: "Context", Documentation: "Package context defines operations carried out by a request."},
		{Label: "errors", Kind: "package", Detail: "Error handling", Documentation: "Package errors implements functions to manipulate errors."},
		{Label: "reflect", Kind: "package", Detail: "Reflection", Documentation: "Package reflect implements run-time reflection."},
	}

	return packages
}

// GetSelectorCompletions returns completions for a selector expression
func (c *ContextAnalyzer) GetSelectorCompletions(scope, partial string) []CompletionItem {
	completions := []CompletionItem{}

	// Standard library package members
	switch scope {
	case "fmt":
		completions = append(completions, []CompletionItem{
			{Label: "Print", Kind: "function", Detail: "Print prints arguments to standard output"},
			{Label: "Printf", Kind: "function", Detail: "Printf formats according to a format specifier and writes to standard output"},
			{Label: "Println", Kind: "function", Detail: "Println prints arguments to standard output with spaces"},
			{Label: "Sprintf", Kind: "function", Detail: "Sprintf formats according to a format specifier and returns the resulting string"},
			{Label: "Errorf", Kind: "function", Detail: "Errorf formats according to a format specifier and returns the error"},
		}...)

	case "os":
		completions = append(completions, []CompletionItem{
			{Label: "Exit", Kind: "function", Detail: "Exit causes the current program to exit"},
			{Label: "Getenv", Kind: "function", Detail: "Getenv retrieves the value of the environment variable"},
			{Label: "Setenv", Kind: "function", Detail: "Setenv sets the value of the environment variable"},
			{Label: "Args", Kind: "variable", Detail: "Args hold the command-line arguments"},
			{Label: "Stdin", Kind: "variable", Detail: "Stdin is standard input"},
			{Label: "Stdout", Kind: "variable", Detail: "Stdout is standard output"},
			{Label: "Stderr", Kind: "variable", Detail: "Stderr is standard error"},
		}...)

	case "strings":
		completions = append(completions, []CompletionItem{
			{Label: "Contains", Kind: "function", Detail: "Contains reports whether substr is within s"},
			{Label: "HasPrefix", Kind: "function", Detail: "HasPrefix tests whether the string s begins with prefix"},
			{Label: "HasSuffix", Kind: "function", Detail: "HasSuffix tests whether the string s ends with suffix"},
			{Label: "Index", Kind: "function", Detail: "Index returns the index of the first instance of substr in s"},
			{Label: "Join", Kind: "function", Detail: "Join concatenates the elements of a to create a single string"},
			{Label: "Split", Kind: "function", Detail: "Split slices s into all substrings separated by sep"},
			{Label: "ToLower", Kind: "function", Detail: "ToLower returns s with all Unicode letters mapped to their lower case"},
			{Label: "ToUpper", Kind: "function", Detail: "ToUpper returns s with all Unicode letters mapped to their upper case"},
		}...)
	}

	// Filter by partial match
	var filtered []CompletionItem
	for _, item := range completions {
		if strings.HasPrefix(item.Label, partial) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// GetVariableCompletions returns common variable completions
func (c *ContextAnalyzer) GetVariableCompletions(partial string) []CompletionItem {
	variables := []CompletionItem{
		{"err", "variable", "Error variable", "Standard error variable"},
		{"result", "variable", "Result variable", "Function result variable"},
		{"value", "variable", "Value variable", "Generic value variable"},
		{"data", "variable", "Data variable", "Data variable"},
		{"i", "variable", "Loop counter", "Loop index variable"},
		{"n", "variable", "Count variable", "Count or length variable"},
	}

	var filtered []CompletionItem
	for _, item := range variables {
		if strings.HasPrefix(item.Label, partial) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// GetFunctionCompletions returns common function completions
func (c *ContextAnalyzer) GetFunctionCompletions(partial string) []CompletionItem {
	functions := []CompletionItem{
		{"main", "function", "Main function", "Program entry point"},
		{"init", "function", "Init function", "Package initialization"},
		{"New", "function", "Constructor function", "Constructor pattern"},
		{"Get", "function", "Getter function", "Getter pattern"},
		{"Set", "function", "Setter function", "Setter pattern"},
		{"Is", "function", "Predicate function", "Boolean check pattern"},
		{"Handle", "function", "Handler function", "Event handler pattern"},
		{"Process", "function", "Process function", "Data processing function"},
	}

	var filtered []CompletionItem
	for _, item := range functions {
		if strings.HasPrefix(item.Label, partial) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// GetCompletionSuggestions returns general completion suggestions
func (c *ContextAnalyzer) GetCompletionSuggestions(ctx CompletionContext) []CompletionItem {
	var suggestions []CompletionItem

	switch ctx.Type {
	case ContextPackageImport:
		suggestions = c.GetStandardPackages()
	case ContextSelector:
		suggestions = c.GetSelectorCompletions(ctx.Scope, ctx.Prefix)
	case ContextVariableDeclaration:
		suggestions = c.GetVariableCompletions(ctx.Prefix)
	case ContextFunctionCall:
		suggestions = c.GetFunctionCompletions(ctx.Prefix)
	case ContextTypeDeclaration:
		suggestions = []CompletionItem{
			{"string", "type", "String type", "String data type"},
			{"int", "type", "Integer type", "Integer data type"},
			{"bool", "type", "Boolean type", "Boolean data type"},
			{"float64", "type", "Float type", "64-bit float type"},
			{"[]string", "type", "String slice", "Slice of strings"},
			{"[]int", "type", "Integer slice", "Slice of integers"},
			{"map[string]interface{}", "type", "Generic map", "Map with string keys and interface values"},
			{"interface{}", "type", "Interface type", "Empty interface"},
		}
	default:
		// General completions - combine common patterns
		suggestions = append(suggestions, c.GetFunctionCompletions(ctx.Prefix)...)
		suggestions = append(suggestions, c.GetVariableCompletions(ctx.Prefix)...)
	}

	return suggestions
}
