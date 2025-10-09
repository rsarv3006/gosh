//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
)

func RunREPL(state *ShellState, evaluator *GoEvaluator, spawner *ProcessSpawner, builtins *BuiltinHandler) error {
	router := NewRouter(builtins)

	// Setup readline with multiline support
	rl, err := readline.NewEx(&readline.Config{
		AutoComplete: NewGoshCompleter(),
		UniqueEditLine: true, // Enable better completion
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
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
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
			if _, found := FindInPath(command); !found {
				result = ExecutionResult{
					Output:   fmt.Sprintf("gosh: command not found: %s", command),
					ExitCode: 127,
					Error:    fmt.Errorf("command not found: %s", command),
				}
			} else {
				result = spawner.ExecuteInteractive(command, args)
			}
		}

		// Display output
		if result.Output != "" {
			fmt.Println(result.Output)
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
	if strings.HasSuffix(input, ",") || 
	   strings.HasSuffix(input, "+") || 
	   strings.HasSuffix(input, "-") || 
	   strings.HasSuffix(input, "*") || 
	   strings.HasSuffix(input, "/") ||
	   strings.HasSuffix(input, "||") ||
	   strings.HasSuffix(input, "&&") {
		return false
	}
	
	return true
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
