//go:build darwin || linux

package main

import (
	"strings"
	"unicode"
)

type Router struct {
	builtins *BuiltinHandler
}

func NewRouter(builtins *BuiltinHandler) *Router {
	return &Router{builtins: builtins}
}

// Route determines what to do with the input
func (r *Router) Route(input string) (InputType, string, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return InputTypeCommand, "", nil
	}

	// Parse into command and args
	command, args := r.parseInput(input)

	// Check builtins first
	if r.builtins.IsBuiltin(command) {
		return InputTypeBuiltin, command, args
	}

	// Check for Go syntax markers
	if r.looksLikeGo(input) {
		return InputTypeGo, input, nil
	}

	// Default to command execution
	return InputTypeCommand, command, args
}

func (r *Router) looksLikeGo(input string) bool {
	input = strings.TrimSpace(input)

	// Go keywords that indicate code
	goKeywords := []string{
		"var ", "const ", "func ", "type ", "struct ", "interface ",
		"import ", "package ", "for ", "range ", "if ", "switch ",
		"return ", "go ", "defer ", "select ", "case ",
	}

	for _, kw := range goKeywords {
		if strings.HasPrefix(input, kw) {
			return true
		}
	}

	// Check for assignment operators
	if strings.Contains(input, ":=") || strings.Contains(input, "=") {
		// But not ==, !=, <=, >=
		if !strings.Contains(input, "==") && !strings.Contains(input, "!=") &&
			!strings.Contains(input, "<=") && !strings.Contains(input, ">=") {
			return true
		}
	}

	// Check for arithmetic or comparison operators (likely Go expression)
	if strings.ContainsAny(input, "+-*/%<>!&|^") {
		return true
	}

	// Check for function calls with string literals (common in Go)
	if strings.Contains(input, `"`) && strings.Contains(input, "(") {
		return true
	}

	// Common Go functions
	goFunctions := []string{
		"fmt.Print", "fmt.Sprint", "len(", "cap(", "make(",
		"append(", "copy(", "delete(", "panic(", "recover(",
	}

	for _, fn := range goFunctions {
		if strings.Contains(input, fn) {
			return true
		}
	}

	// Check for multi-line or block structure
	if strings.Contains(input, "{") || strings.Contains(input, "}") {
		return true
	}

	// Single word with no special shell chars - could be a variable reference
	// But need to distinguish from commands
	if !strings.ContainsAny(input, " \t/.-") {
		// Single identifier - try as Go variable if it's not a known command
		// This handles typos by letting them fail as commands
		if _, found := FindInPath(input); found {
			return false // It's a command
		}
		// Check if it contains only valid Go identifier characters
		// If it looks like a filename or typo, let it fail as a command
		hasValidGoChars := true
		for _, ch := range input {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
				hasValidGoChars = false
				break
			}
		}
		return hasValidGoChars
	}

	return false
}

func (r *Router) parseInput(input string) (string, []string) {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for i, char := range input {
		switch {
		case (char == '"' || char == '\'') && (i == 0 || input[i-1] != '\\'):
			if !inQuote {
				inQuote = true
				quoteChar = char
			} else if char == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				current.WriteRune(char)
			}
		case unicode.IsSpace(char) && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if len(args) == 0 {
		return "", nil
	}

	return args[0], args[1:]
}

