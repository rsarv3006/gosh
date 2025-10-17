//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ProcessSpawner struct {
	state *ShellState
}

func NewProcessSpawner(state *ShellState) *ProcessSpawner {
	return &ProcessSpawner{state: state}
}

func (p *ProcessSpawner) Execute(command string, args []string) ExecutionResult {
	var cmd *exec.Cmd

	isGitStatus := (command == "git" && len(args) > 0 && args[0] == "status") ||
		(command == "env" && len(args) >= 2 && args[len(args)-1] == "status" && args[len(args)-2] == "git")

	if isGitStatus {
		env := p.state.EnvironmentSlice()

		if command == "env" {
			cmd = exec.Command("git", "status")
			for _, arg := range args {
				if strings.HasPrefix(arg, "GIT_COLOR=") || strings.HasPrefix(arg, "TERM=") ||
					strings.HasPrefix(arg, "CLICOLOR=") || strings.HasPrefix(arg, "CLICOLOR_FORCE=") {
					env = append(env, arg)
				}
			}
			if !containsEnv(env, "GIT_COLOR") {
				env = append(env, "GIT_COLOR=always")
			}
			if !containsEnv(env, "TERM") {
				env = append(env, "TERM=xterm-256color")
			}
			if !containsEnv(env, "CLICOLOR") {
				env = append(env, "CLICOLOR=1")
			}
			if !containsEnv(env, "CLICOLOR_FORCE") {
				env = append(env, "CLICOLOR_FORCE=1")
			}
		} else {
			env = append(env, "GIT_COLOR=always", "TERM=xterm-256color", "CLICOLOR=1", "CLICOLOR_FORCE=1")
			cmd = exec.Command(command, args...)
		}

		cmd.Dir = p.state.WorkingDirectory
		cmd.Env = env
		cmd.Stdin = os.Stdin
	} else if command == "ls" {
		env := p.state.EnvironmentSlice()
		env = append(env, "CLICOLOR=1", "CLICOLOR_FORCE=1", "TERM=xterm-256color")
		cmd = exec.Command(command, args...)
		cmd.Dir = p.state.WorkingDirectory
		cmd.Env = env
		cmd.Stdin = os.Stdin
	} else if wantsColorForCommand(command, args) {
		env := p.state.EnvironmentSlice()
		env = append(env, "CLICOLOR=1", "CLICOLOR_FORCE=1", "TERM=xterm-256color", "FORCE_COLOR=1")

		if command == "git" {
			if !containsEnv(env, "GIT_COLOR") {
				env = append(env, "GIT_COLOR=always")
			}
		}

		cmd = exec.Command(command, args...)
		cmd.Dir = p.state.WorkingDirectory
		cmd.Env = env
		cmd.Stdin = os.Stdin
	} else {
		cmd = exec.Command(command, args...)
		cmd.Dir = p.state.WorkingDirectory
		cmd.Env = p.state.EnvironmentSlice()
		cmd.Stdin = os.Stdin
	}

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

func (p *ProcessSpawner) ExecuteInteractive(command string, args []string) ExecutionResult {
	commandPath := command
	// Don't resolve local paths through PATH - use them directly
	if !strings.HasPrefix(command, "./") {
		if fullPath, found := FindInPath(command, p.state.Environment["PATH"]); found {
			commandPath = fullPath
		}
	}

	env := p.state.EnvironmentSlice()
	var modifiedArgs []string

	if wantsColorForCommand(command, args) {
		if !containsEnv(env, "CLICOLOR") {
			env = append(env, "CLICOLOR=1")
		}
		if !containsEnv(env, "CLICOLOR_FORCE") {
			env = append(env, "CLICOLOR_FORCE=1")
		}
		if !containsEnv(env, "TERM") {
			env = append(env, "TERM=xterm-256color")
		}
		if !containsEnv(env, "FORCE_COLOR") {
			env = append(env, "FORCE_COLOR=1")
		}

		if command == "ls" {
			modifiedArgs = make([]string, len(args))
			copy(modifiedArgs, args)
			hasColorFlag := false
			for _, arg := range args {
				if arg == "--color=always" || arg == "--color" || strings.HasPrefix(arg, "--color=") {
					hasColorFlag = true
					break
				}
			}
			if !hasColorFlag {
				modifiedArgs = append(modifiedArgs, "--color=always")
			}
		}

		if command == "git" {
			if !containsEnv(env, "GIT_COLOR") {
				env = append(env, "GIT_COLOR=always")
			}
		}
	}

	var finalArgs []string
	if len(modifiedArgs) > 0 {
		finalArgs = make([]string, len(modifiedArgs))
		copy(finalArgs, modifiedArgs)
		for i, arg := range finalArgs {
			finalArgs[i] = p.expandShellVariables(arg)
		}
	} else {
		finalArgs = make([]string, len(args))
		copy(finalArgs, args)
		for i, arg := range finalArgs {
			finalArgs[i] = p.expandShellVariables(arg)
		}
	}

	cmd := exec.Command(commandPath, finalArgs...)
	cmd.Dir = p.state.WorkingDirectory
	cmd.Env = env
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

func (p *ProcessSpawner) expandShellVariables(input string) string {
	result := input

	for key, value := range p.state.Environment {
		varPattern := "$" + key
		if strings.Contains(result, varPattern) {
			result = strings.ReplaceAll(result, varPattern, value)
		}
	}

	return result
}

func FindInPath(command string, pathEnv string) (string, bool) {
	if pathEnv == "" {
		pathEnv = "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	}

	for _, dir := range strings.Split(pathEnv, ":") {
		if dir == "" {
			continue
		}
		fullPath := filepath.Join(dir, command)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}

	return "", false
}

func containsEnv(env []string, key string) bool {
	for _, v := range env {
		if strings.HasPrefix(v, key+"=") {
			return true
		}
	}
	return false
}

func wantsColorForCommand(command string, args []string) bool {
	coloredCommands := map[string]bool{
		"ls":     true,
		"grep":   true,
		"git":    true,
		"docker": true,
		"npm":    true,
		"yarn":   true,
		"node":   true,
		"python": true,
		"pip":    true,
		"curl":   true,
		"wget":   true,
		"bat":    true,
		"exa":    true,
		"rg":     true,
		"fd":     true,
		"tree":   true,
	}

	if coloredCommands[command] {
		return true
	}

	if command == "git" && len(args) > 0 {
		gitColoredSubcommands := map[string]bool{
			"status": true,
			"diff":   true,
			"log":    true,
			"show":   true,
			"branch": true,
			"stash":  true,
			"blame":  true,
		}
		return gitColoredSubcommands[args[0]]
	}

	if command == "grep" || command == "rg" || command == "ls" {
		return true
	}

	return false
}
