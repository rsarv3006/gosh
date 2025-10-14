//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "v0.0.5"

func main() {
	// Check for command line flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("gosh %s\n", version)
			os.Exit(0)
		case "-h", "--help":
			fmt.Printf("gosh %s - Go shell with yaegi\n\n", version)
			fmt.Println("Usage:")
			fmt.Println("  gosh          Start the gosh interactive shell")
			fmt.Println("  gosh --version Show version information")
			fmt.Println("  gosh --help    Show this help message")
			fmt.Println("\nFlags:")
			fmt.Println("  -v, --version  Show version information")
			fmt.Println("  -h, --help     Show this help message")
			fmt.Println("\nVisit https://github.com/rsarv3006/gosh for more information.")
			os.Exit(0)
		case "-c":
			// Execute single command and exit
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: gosh -c '<command>'\n")
				os.Exit(1)
			}
			command := strings.Join(os.Args[2:], " ")
			state := NewShellState()
			evaluator := NewGoEvaluator()
			spawner := NewProcessSpawner(state)
			
			evaluator.SetupWithShell(state, spawner)
			
			// Use evaluator to execute the command
			result := evaluator.Eval(command)
			fmt.Print(result.Output)
			os.Exit(result.ExitCode)
		}
	}

	state := NewShellState()
	evaluator := NewGoEvaluator()
	spawner := NewProcessSpawner(state)
	builtins := NewBuiltinHandler(state)
	colors := GetColorManager()

	// Setup evaluator with shell access
	evaluator.SetupWithShell(state, spawner)

	fmt.Println(colors.StyleMessage(fmt.Sprintf("gosh %s - Go shell with yaegi", version), "welcome"))
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
