//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"strings"
)

// Shell API functions available to user config

// RunShell executes a command and returns its output as a string with command substitution
func RunShell(name string, args ...string) (string, error) {
	// Return a command substitution string that will be processed by the evaluator
	var cmdStr strings.Builder
	cmdStr.WriteString("$(")
	cmdStr.WriteString(name)
	for _, arg := range args {
		cmdStr.WriteString(" ")
		cmdStr.WriteString(arg)
	}
	cmdStr.WriteString(")")
	
	// Debug: return the raw string for testing
	return cmdStr.String(), nil
}

// ExecShell executes a command and connects it directly to stdin/stdout/stderr
func ExecShell(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ShellCmd is an alias for RunShell for convenience
func ShellCmd(command string, args ...string) (string, error) {
	return RunShell(command, args...)
}
