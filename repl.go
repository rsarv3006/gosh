//go:build darwin || linux

package main

import (
	"bufio"
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

	// Setup signal handling
	setupSignals(state)

	// Try readline first, fallback to basic mode if it fails
	rl, useReadline := setupReadlineWithFallback()
	if useReadline {
		defer rl.Close()
	} else {
		fmt.Fprintln(os.Stderr, "\n🚨 Readline unavailable, using basic mode. Arrow keys and tab completion disabled.")
		fmt.Fprint(os.Stderr, "Check your terminal (TERM=$TERM) or ~/.inputrc configuration.\n")
	}

	for !state.ShouldExit {
		var input string
		var err error

		if useReadline {
			// Use enhanced readline mode
			rl.SetPrompt(state.GetPrompt())
			input, err = rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt {
					continue
				}
				if err == io.EOF {
					break
				}
				// Readline error - try to fall back
				fmt.Fprintln(os.Stderr, "\n🚨 Readline error, switching to basic mode...")
				useReadline = false
				continue
			}
		} else {
			// Use basic stdin mode
			fmt.Print(state.GetPrompt())
			reader := bufio.NewReader(os.Stdin)
			input, err = reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintln(os.Stderr, "\n🚨 Input error, continuing...")
				continue
			}
			input = strings.TrimSuffix(input, "\n")
		}

		// Skip empty lines
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Handle multiline input accumulation
		if useReadline {
			for !isComplete(input) {
				rl.SetPrompt("... ")
				line, err := rl.Readline()
				if err != nil {
					break
				}
				input += "\n" + line
			}
		} else {
			// Basic multiline support without fancy prompts
			for !isComplete(input) {
				fmt.Print("... ")
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				input += "\n" + strings.TrimSuffix(line, "\n")
			}
		}

		// Route and execute with recovery
		result := routeAndExecuteWithRecovery(router, evaluator, spawner, builtins, input, state)

		// Display output with colors
		if result.Output != "" {
			colors := GetColorManager()
			if result.ExitCode != 0 {
				// Error output
				fmt.Println(colors.StyleOutput(result.Output, "error"))
			} else {
				// Success output - write raw bytes to preserve tabs and formatting exactly
				os.Stdout.WriteString(result.Output)
				if !strings.HasSuffix(result.Output, "\n") {
					os.Stdout.WriteString("\n")
				}
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

// setupReadlineWithFallback attempts to setup readline with graceful fallback
func setupReadlineWithFallback() (*readline.Instance, bool) {
	rl, err := readline.NewEx(&readline.Config{
		AutoComplete: NewGoshCompleter(),
	})
	if err != nil {
		return nil, false
	}
	return rl, true
}

// routeAndExecuteWithRecovery adds panic recovery for safe execution
func routeAndExecuteWithRecovery(router *Router, evaluator *GoEvaluator, spawner *ProcessSpawner, builtins *BuiltinHandler, input string, state *ShellState) ExecutionResult {
	// Recover from panics during execution
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n🚨 Panic recovered: %v\n", r)
			fmt.Fprintln(os.Stderr, "Type 'exit' to quit or continue with a new command.")
		}
	}()

	inputType, command, args := router.Route(input)

	switch inputType {
	case InputTypeBuiltin:
		return builtins.Execute(command, args)

	case InputTypeGo:
		// Add recovery for yaegi crashes
		return evaluator.EvalWithRecovery(input)

	case InputTypeCommand:
		// Check if command exists
		if _, found := FindInPath(command, state.Environment["PATH"]); !found {
			return ExecutionResult{
				Output:   fmt.Sprintf("gosh: command not found: %s", command),
				ExitCode: 127,
				Error:    fmt.Errorf("command not found: %s", command),
			}
		} else {
			return spawner.ExecuteInteractive(command, args)
		}
	}

	return ExecutionResult{ExitCode: 0}
}

func setupSignals(state *ShellState) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "🚨 Signal handler panic recovered: %v\n", r)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "🚨 Signal goroutine panic recovered: %v\n", r)
			}
		}()

		for sig := range sigChan {
			switch sig {
			case os.Interrupt:
				// Ctrl+C - interrupt current process or print newline
				if state.CurrentProcess != nil {
					if err := state.CurrentProcess.Signal(os.Interrupt); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to signal process: %v\n", err)
					}
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
