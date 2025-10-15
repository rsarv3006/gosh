//go:build darwin || linux

package main

import (
	"strings"
	"unicode"
)

type Router struct {
	builtins *BuiltinHandler
	state    *ShellState
}

func NewRouter(builtins *BuiltinHandler, state *ShellState) *Router {
	return &Router{builtins: builtins, state: state}
}

// Route determines what to do with the input
func (r *Router) Route(input string) (InputType, string, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return InputTypeCommand, "", nil
	}

	// Parse into command and args
	command, args := r.parseInput(input)

	

	// Check for command substitution $(command)
	if r.hasCommandSubstitution(input) {
		return InputTypeGo, input, nil
	}

	// Check builtins first
	if r.builtins.IsBuiltin(command) {
		return InputTypeBuiltin, command, args
	}

	// Check for Go syntax patterns BEFORE shell command check
	// This is critical - keywords like 'func' should NOT be treated as shell commands
	if r.looksLikeGoCode(input) {
		return InputTypeGo, input, nil
	}

	// Check if it looks like a shell command
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
	if _, found := FindInPath(command, r.state.Environment["PATH"]); found {
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
	
	// If it looks like a shell command but is NOT in PATH, let it fall back to Go
	return false
}

// looksLikeGoCode checks if the input looks like Go code that should be evaluated
func (r *Router) looksLikeGoCode(input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return false
	}

	// Check for Go-specific patterns that clearly indicate Go code
	
	// Function definitions
	if strings.HasPrefix(input, "func ") {
		return true
	}
	
	// Go keywords that indicate Go code (excluding when used as shell commands)
	goKeywords := []string{
		"var ", "const ", "type ", "import ", "package ",
		"if ", "else", "for ", "switch ", "select ", "case ", "default ",
		"defer ", "return ", "break ", "continue ", "fallthrough ",
		"struct ", "interface ", "map ", "chan ",
		"go ", // for 'go func(){}', 'go fmt.Println()', etc.
	}
	
	for _, keyword := range goKeywords {
		if strings.HasPrefix(input, keyword) {
			// Special case: 'go' should only trigger if it's not the first word (not a command)
			if keyword == "go " {
				words := strings.Fields(input)
				if len(words) > 0 && words[0] == "go" {
					// 'go' is the first word, treat as shell command
					continue
				}
			}
			return true
		}
	}
	
	// Check if input contains Go syntax patterns that aren't typical shell
	// (but be careful not to over-match shell commands that use similar syntax)
	
	// Type declarations with Go syntax
	if strings.Contains(input, ":=") && !strings.Contains(input, "$(") {
		return true
	}
	
	// Go-specific syntax patterns
	if strings.ContainsAny(input, "{}()") && 
	   !strings.Contains(input, "|") && 
	   !strings.Contains(input, ">") && 
	   !strings.Contains(input, "<") {
		// Contains Go braces/parentheses but not typical shell operators
		return true
	}
	
	// Go types (common patterns)
	goTypes := []string{
		" string ", " int ", " bool ", " float64 ", " float32 ",
		" byte ", " rune ", " error ", " interface{} ",
		" int8 ", " int16 ", " int32 ", " int64 ",
		" uint8 ", " uint16 ", " uint32 ", " uint64 ",
	}
	
	for _, goType := range goTypes {
		if strings.Contains(input, goType) {
			return true
		}
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

