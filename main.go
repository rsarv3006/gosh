//go:build darwin || linux

package main

import (
	"fmt"
	"os"
)

func main() {
	state := NewShellState()
	evaluator := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	builtins := NewBuiltinHandler(state)

	// Setup evaluator with shell access
	evaluator.SetupWithShell(state, spawner)

	fmt.Println("gosh v0.0.1 - Go shell with yaegi")
	fmt.Println("Type 'exit' to quit, try some Go code or shell commands!")
	fmt.Println()

	// Load config.go if it exists
	if err := evaluator.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config loading error: %v\n", err)
	}

	if err := RunREPL(state, evaluator, spawner, builtins); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(state.ExitCode)
}
