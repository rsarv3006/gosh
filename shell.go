//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"strings"
)

func RunShell(name string, args ...string) (string, error) {
	var cmdStr strings.Builder
	cmdStr.WriteString("$(")
	cmdStr.WriteString(name)
	for _, arg := range args {
		cmdStr.WriteString(" ")
		cmdStr.WriteString(arg)
	}
	cmdStr.WriteString(")")

	return cmdStr.String(), nil
}

func ExecShell(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
