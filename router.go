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

	// Check for command substitution $(command)
	if r.hasCommandSubstitution(input) {
		return InputTypeGo, input, nil
	}

	// Parse into command and args
	command, args := r.parseInput(input)

	// Check builtins first
	if r.builtins.IsBuiltin(command) {
		return InputTypeBuiltin, command, args
	}

	// Check if it looks like a shell command first
	if r.looksLikeShellCommand(input) {
		return InputTypeCommand, command, args
	}

	// Default to Go evaluation - safer fallback
	return InputTypeGo, input, nil
}

// hasCommandSubstitution checks for $(command) syntax
func (r *Router) hasCommandSubstitution(input string) bool {
	start := strings.Index(input, "$(")
	if start == -1 {
		return false
	}
	
	// Find matching closing parenthesis
	for i := start + 2; i < len(input); i++ {
		if input[i] == '(' {
			// Nested parentheses - find closing for this level
			depth := 1
			for j := i + 1; j < len(input) && depth > 0; j++ {
				if input[j] == '(' {
					depth++
				} else if input[j] == ')' {
					depth--
				}
			}
			if depth > 0 {
				return false // Unbalanced parentheses
			}
			i += depth * 2 // Skip past nested parentheses
		} else if input[i] == ')' {
			return true // Found matching closing parenthesis
		}
	}
	return false // No matching closing parenthesis found
}



func (r *Router) looksLikeShellCommand(input string) bool {
	input = strings.TrimSpace(input)
	
	// Empty string is definitely not a shell command
	if input == "" {
		return false
	}

	// Check for obvious shell patterns
	command, args := r.parseInput(input)
	
	// If first word is in PATH, it's definitely a command
	if _, found := FindInPath(command); found {
		return true
	}
	
	// Shell command patterns:
	
	// Has arguments/flags (dash or slash patterns)
	if len(args) > 0 {
		for _, arg := range args {
			// Shell flags typically start with -
			if strings.HasPrefix(arg, "-") {
				return true
			}
			// Shell paths often contain /
			if strings.Contains(arg, "/") {
				return true
			}
		}
	}
	
	// Contains shell operators like pipes, redirects
	if strings.ContainsAny(input, "|><") {
		return true
	}
	
	// Go syntax patterns - if detected, definitely NOT a shell command
	if strings.ContainsAny(input, "{}();:=") || strings.Contains(input, "\"") {
		return false // These are Go patterns
	}
	
	// Single word with no obvious Go syntax, try as command first
	if !strings.ContainsAny(input, "|><") {
		return true // Try as command first, fallback to Go if not in PATH
	}
	
	// Default to false (let it fall back to Go)
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

