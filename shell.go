//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"os/exec"
)

// Shell API functions available to user config

// RunShell executes a command and returns its output as a string
func RunShell(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
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
