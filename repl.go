//go:build darwin || linux

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
)

func RunREPL(state *ShellState, evaluator *GoEvaluator, spawner *ProcessSpawner, builtins *BuiltinHandler) error {
	router := NewRouter(builtins, state)

	// Setup readline with multiline support
	rl, err := readline.NewEx(&readline.Config{
		AutoComplete: NewGoshCompleter(),
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	// Setup signal handling
	setupSignals(state)

	for !state.ShouldExit {
		// Set prompt
		rl.SetPrompt(state.GetPrompt())

		// Read input with multiline support
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			if err == io.EOF {
				break
			}
			return err
		}

		// Skip empty lines
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Handle multiline input accumulation
		for !isComplete(input) {
			// Set continuation prompt
			rl.SetPrompt("... ")

			// Read next line
			line, err := rl.Readline()
			if err != nil {
				break
			}

			// Add continue marker for readability
			input += "\n" + line
		}

		// Route and execute
		inputType, command, args := router.Route(input)

		var result ExecutionResult

		switch inputType {
		case InputTypeBuiltin:
			result = builtins.Execute(command, args)

		case InputTypeGo:
			result = evaluator.Eval(input)

		case InputTypeCommand:
			// Check if command exists
			if _, found := FindInPath(command, state.Environment["PATH"]); !found {
				result = ExecutionResult{
					Output:   fmt.Sprintf("gosh: command not found: %s", command),
					ExitCode: 127,
					Error:    fmt.Errorf("command not found: %s", command),
				}
			} else {
				result = spawner.ExecuteInteractive(command, args)
			}
		}

		// Display output with colors
		if result.Output != "" {
			colors := GetColorManager()
			if result.ExitCode != 0 {
				// Error output
				fmt.Println(colors.StyleOutput(result.Output, "error"))
			} else {
				// Success output - but don't color Go evaluation results that might interfere
				fmt.Println(result.Output)
			}
		}

		// Update last exit code (could store this in state if needed)
		if result.ExitCode != 0 && result.Error != nil {
			// Optionally print error info
		}
	}

	return nil
}

func isComplete(input string) bool {
	input = strings.TrimSpace(input)

	// Check for unclosed braces
	openBraces := strings.Count(input, "{")
	closeBraces := strings.Count(input, "}")
	if openBraces != closeBraces {
		return false
	}

	// Check for unclosed parentheses
	openParens := strings.Count(input, "(")
	closeParens := strings.Count(input, ")")
	if openParens != closeParens {
		return false
	}

	// Check for unclosed brackets
	openBrackets := strings.Count(input, "[")
	closeBrackets := strings.Count(input, "]")
	if openBrackets != closeBrackets {
		return false
	}

	// Check if line ends with incomplete statement
	// But be more careful about "/" - it could be path completion
	if strings.HasSuffix(input, ",") ||
		strings.HasSuffix(input, "+") ||
		strings.HasSuffix(input, "-") ||
		strings.HasSuffix(input, "*") ||
		(strings.HasSuffix(input, "/") && !looksLikePathCompletion(input)) ||
		strings.HasSuffix(input, "||") ||
		strings.HasSuffix(input, "&&") {
		return false
	}

	return true
}

// looksLikePathCompletion checks if the trailing "/" is likely from path completion
func looksLikePathCompletion(input string) bool {
	input = strings.TrimSpace(input)
	
	// If it ends with "/" and looks like a path command, treat it as complete
	if !strings.HasSuffix(input, "/") {
		return false
	}
	
	// Split into words and check if it looks like a command with path argument
	words := strings.Fields(input)
	if len(words) == 0 {
		return false
	}
	
	command := words[0]
	lastWord := words[len(words)-1]
	
	// Common commands that take directory paths
	pathCommands := map[string]bool{
		"cd": true, "ls": true, "pwd": false, "cat": false, 
		"grep": false, "find": true, "mkdir": true, "rmdir": true,
		"rm": false, "mv": false, "cp": false, "touch": false,
	}
	
	// If it's a known path command and the last word ends with "/", it's path completion
	if pathCommands[command] && strings.HasSuffix(lastWord, "/") {
		return true
	}
	
	// If the last word contains path separators, it's likely a path
	if strings.Contains(lastWord, "/") || strings.HasPrefix(lastWord, "~") {
		return true
	}
	
	// Heuristic: if there's only one word that ends with "/" and no Go syntax, it's likely a path
	if len(words) == 1 && !strings.ContainsAny(input, "{}();:=") {
		return true
	}
	
	return false
}

func setupSignals(state *ShellState) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case os.Interrupt:
				// Ctrl+C - interrupt current process or print newline
				if state.CurrentProcess != nil {
					state.CurrentProcess.Signal(os.Interrupt)
					fmt.Println("^C")
				} else {
					fmt.Println("^C")
				}
			case syscall.SIGTERM:
				// Graceful shutdown
				fmt.Println("\nShutting down...")
				os.Exit(0)
			}
		}
	}()
}
