//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"os/exec"
)

type ProcessSpawner struct {
	state *ShellState
}

func NewProcessSpawner(state *ShellState) *ProcessSpawner {
	return &ProcessSpawner{state: state}
}

// Execute runs a command and returns the result
func (p *ProcessSpawner) Execute(command string, args []string) ExecutionResult {
	cmd := exec.Command(command, args...)
	cmd.Dir = p.state.WorkingDirectory
	cmd.Env = p.state.EnvironmentSlice()
	cmd.Stdin = os.Stdin

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	return ExecutionResult{
		Output:   output,
		ExitCode: exitCode,
		Error:    err,
	}
}

// ExecuteInteractive runs a command with direct terminal access
func (p *ProcessSpawner) ExecuteInteractive(command string, args []string) ExecutionResult {
	cmd := exec.Command(command, args...)
	cmd.Dir = p.state.WorkingDirectory
	cmd.Env = p.state.EnvironmentSlice()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return ExecutionResult{
			Output:   "",
			ExitCode: 1,
			Error:    err,
		}
	}

	// Track the current process
	p.state.CurrentProcess = cmd.Process
	defer func() {
		p.state.CurrentProcess = nil
	}()

	err = cmd.Wait()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecutionResult{
		Output:   "",
		ExitCode: exitCode,
		Error:    err,
	}
}

// FindInPath checks if a command exists in PATH
func FindInPath(command string) (string, bool) {
	path, err := exec.LookPath(command)
	return path, err == nil
}
