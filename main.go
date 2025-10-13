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
	colors := GetColorManager()

	// Setup evaluator with shell access
	evaluator.SetupWithShell(state, spawner)

	fmt.Println(colors.StyleMessage("gosh v0.0.2 - Go shell with yaegi", "welcome"))
	fmt.Println(colors.StyleMessage("Type 'exit' to quit, try some Go code or shell commands!", "welcome"))
	fmt.Println()

	// Load config.go if it exists
	if err := evaluator.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", colors.StyleOutput(fmt.Sprintf("Config loading error: %v", err), "error"))
	}

	if err := RunREPL(state, evaluator, spawner, builtins); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", colors.StyleOutput(fmt.Sprintf("Error: %v", err), "error"))
		os.Exit(1)
	}

	os.Exit(state.ExitCode)
}
