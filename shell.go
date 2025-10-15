//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// Shell API functions available to user config

// RunShell executes a command and returns its output as a string
func RunShell(name string, args ...string) (string, error) {
	// This won't work without access to global state
	// For now, revert to the simple implementation
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	
	// Special handling for git status with colors
	if name == "git" && len(args) > 0 && args[0] == "status" {
		cmd.Env = append(os.Environ(), "GIT_COLOR=always", "TERM=xterm-256color")
		} else if name == "env" && len(args) >= 3 && args[len(args)-2] == "git" && args[len(args)-1] == "status" {
		// Convert env GIT_COLOR=always git status to git status with env
		cmd = exec.Command("git", "status")
		envVars := []string{}
		for _, arg := range args[:len(args)-2] {
			if strings.HasPrefix(arg, "GIT_COLOR=") || strings.HasPrefix(arg, "TERM=") {
				envVars = append(envVars, arg)
			}
		}
		envVars = append(envVars, "GIT_COLOR=always", "TERM=xterm-256color")
		cmd.Env = append(os.Environ(), envVars...)
	}
	
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
