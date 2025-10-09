//go:build darwin || linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func RunREPL(state *ShellState, evaluator *GoEvaluator, spawner *ProcessSpawner, builtins *BuiltinHandler) error {
	router := NewRouter(builtins)
	reader := bufio.NewReader(os.Stdin)

	// Setup signal handling
	setupSignals()

	for !state.ShouldExit {
		// Display prompt
		fmt.Print(state.GetPrompt())

		// Read input
		input, err := reader.ReadString('\n')
		if err != nil {
			// EOF (Ctrl+D)
			fmt.Println()
			break
		}

		input = input[:len(input)-1] // Remove newline

		// Skip empty lines
		if input == "" {
			continue
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

func setupSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case os.Interrupt:
				// Ctrl+C - print newline for cleanliness
				fmt.Println()
				// Don't exit, just interrupt current operation
			case syscall.SIGTERM:
				// Graceful shutdown
				fmt.Println("\nShutting down...")
				os.Exit(0)
			}
		}
	}()
}
