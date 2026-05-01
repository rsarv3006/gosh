//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
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
			os.Exit(0)
		case "-c":
			if len(os.Args) < 3 {
				fmt.Fprintf(os.Stderr, "Usage: gosh -c '<command>'\n")
				os.Exit(1)
			}
			command := ""
			for _, arg := range os.Args[2:] {
				if command != "" {
					command += " "
				}
				command += arg
			}
			session := NewSessionState()
			evaluator := NewGoEvaluator()
			spawner := NewProcessSpawner(&ShellState{WorkingDirectory: session.WorkingDir})
			builtins := NewBuiltinHandler(&ShellState{WorkingDirectory: session.WorkingDir})

			evaluator.SetupWithShell(&ShellState{WorkingDirectory: session.WorkingDir}, spawner)
			evaluator.SetupWithBuiltins(builtins)

			if err := evaluator.LoadConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Config loading error: %v\n", err)
			}

			if strings.HasPrefix(command, "go> ") {
				command = strings.TrimPrefix(command, "go> ")
				result := evaluator.Eval(command)
				fmt.Print(result.Output)
				os.Exit(result.ExitCode)
			} else {
				parts := strings.Fields(command)
				if len(parts) == 0 {
					os.Exit(0)
				}
				result := spawner.ExecuteInteractive(parts[0], parts[1:])
				fmt.Print(result.Output)
				os.Exit(result.ExitCode)
			}
		}
	}

	session := NewSessionState()
	evaluator := NewGoEvaluator()
	spawner := NewProcessSpawner(&ShellState{WorkingDirectory: session.WorkingDir})
	builtins := NewBuiltinHandler(&ShellState{WorkingDirectory: session.WorkingDir})

	evaluator.SetupWithShell(&ShellState{WorkingDirectory: session.WorkingDir}, spawner)
	evaluator.SetupWithBuiltins(builtins)

	if err := evaluator.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config loading error: %v\n", err)
	}

	p := tea.NewProgram(initialModel(session, evaluator, spawner, builtins), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
