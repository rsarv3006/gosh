//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
)

type ShellState struct {
	WorkingDirectory string
	Environment      map[string]string
	ShouldExit       bool
	ExitCode         int
}

func NewShellState() *ShellState {
	wd, err := os.Getwd()
	if err != nil {
		wd = os.Getenv("HOME")
	}

	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	return &ShellState{
		WorkingDirectory: wd,
		Environment:      env,
		ExitCode:         0,
	}
}

func (s *ShellState) EnvironmentSlice() []string {
	env := make([]string, 0, len(s.Environment))
	for k, v := range s.Environment {
		env = append(env, k+"="+v)
	}
	return env
}

func (s *ShellState) GetPrompt() string {
	dir := s.WorkingDirectory
	home := s.Environment["HOME"]

	if home != "" && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	return dir + "> "
}

// ExpandPath handles ~, environment variables, and path normalization
func (s *ShellState) ExpandPath(path string) string {
	// Handle tilde expansion
	if path == "~" {
		return s.Environment["HOME"]
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(s.Environment["HOME"], path[2:])
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	// Handle relative paths
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.WorkingDirectory, path)
	}

	// Clean the path
	return filepath.Clean(path)
}
