//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("gosh %s\n", GetVersion())
			os.Exit(0)
		case "-h", "--help":
			fmt.Printf("gosh %s - Go shell with yaegi\n\n", GetVersion())
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
			builtins := NewBuiltinHandler(state)

			evaluator.SetupWithShell(state, spawner)
			evaluator.SetupWithBuiltins(builtins)

			// Load config before executing command
			if err := evaluator.LoadConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Config loading error: %v\n", err)
			}

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
	evaluator.SetupWithBuiltins(builtins)

	// Get actual build time from binary modification time
	if exePath, err := os.Executable(); err == nil {
		if info, err := os.Stat(exePath); err == nil {
			buildTime := info.ModTime().Format("2006-01-02 15:04:05")
			fmt.Println(colors.StyleMessage("gosh "+GetVersion()+" - Go shell with yaegi", "welcome") + " (BUILT: " + buildTime + ")")
		} else {
			fmt.Println(colors.StyleMessage("gosh "+GetVersion()+" - Go shell with yaegi", "welcome") + " (BUILT: Unknown)")
		}
	} else {
		fmt.Println(colors.StyleMessage("gosh "+GetVersion()+" - Go shell with yaegi", "welcome") + " (BUILT: Unknown)")
	}
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
