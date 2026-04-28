//go:build darwin || linux

package main

import (
	"strings"
)

// Router is simplified for explicit mode routing.
// In the new architecture, mode is explicit (:go or :sh), so we don't
// need heuristics to guess whether input is Go or shell.
// This router is used only in shell mode to distinguish builtins from commands.
type Router struct {
	builtins *BuiltinHandler
	state    *ShellState
}

func NewRouter(builtins *BuiltinHandler, state *ShellState) *Router {
	return &Router{builtins: builtins, state: state}
}

// Route for shell mode only - determines if input is a builtin or command
func (r *Router) Route(input string) (InputType, string, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return InputTypeCommand, "", nil
	}

	command, args := r.parseInput(input)

	// Check for builtins first
	if r.builtins.IsBuiltin(command) {
		return InputTypeBuiltin, command, args
	}

	// Otherwise treat as shell command
	return InputTypeCommand, command, args
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
		case char == ' ' && !inQuote:
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
