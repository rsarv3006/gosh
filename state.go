//go:build darwin || linux

package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ShellState struct {
	WorkingDirectory string
	Environment      map[string]string
	ShouldExit       bool
	ExitCode         int
	CurrentProcess   *os.Process
	// Cached prompt to avoid expensive color rendering
	cachedPrompt string
	promptHash   string // Content hash to detect changes
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

	state := &ShellState{
		WorkingDirectory: wd,
		Environment:      env,
		ExitCode:         0,
		CurrentProcess:   nil,
	}

	envManager := NewEnvironmentManager(state)
	envManager.InitializeEnvironment()

	return state
}

func (s *ShellState) EnvironmentSlice() []string {
	env := make([]string, 0, len(s.Environment))
	for k, v := range s.Environment {
		env = append(env, k+"="+v)
	}
	return env
}

func (s *ShellState) GetPrompt() string {
	stateHash := s.createPromptHash()

	if s.promptHash == stateHash && s.cachedPrompt != "" {
		return s.cachedPrompt
	}

	newPrompt := s.generatePromptWithColors()

	s.cachedPrompt = newPrompt
	s.promptHash = stateHash

	return newPrompt
}

func (s *ShellState) createPromptHash() string {
	hash := md5.New()
	hash.Write([]byte(s.WorkingDirectory))

	if isInGitRepo(s.WorkingDirectory) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			hash.Write(output)
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (s *ShellState) generatePromptWithColors() string {
	colors := GetColorManager()

	colors.ForceRefresh()
	dir := s.WorkingDirectory
	home := s.Environment["HOME"]

	if home != "" && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	styledDir := colors.StylePrompt(dir, "directory")

	gitBranch := ""
	if isInGitRepo(s.WorkingDirectory) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			branchName := strings.TrimSpace(string(output))

			gitPrefix := colors.StylePrompt("git:", "git_prefix")
			styledBranch := colors.StylePrompt(branchName, "git_branch")
			gitBranch = fmt.Sprintf("%s(%s)", gitPrefix, styledBranch)
		}
	}

	symbol := colors.StylePrompt("> ", "symbol")
	space := colors.StylePrompt(" ", "separator")

	return fmt.Sprintf("%s%s%s%s%s", styledDir, space, gitBranch, space, symbol)
}

func (s *ShellState) ForcePromptRefresh() {
	s.promptHash = ""
}

func (s *ShellState) ExpandPath(path string) string {
	if path == "~" {
		return s.Environment["HOME"]
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(s.Environment["HOME"], path[2:])
	}

	path = os.ExpandEnv(path)

	if !filepath.IsAbs(path) {
		path = filepath.Join(s.WorkingDirectory, path)
	}

	return filepath.Clean(path)
}

func isInGitRepo(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	cmd := exec.Command("git", "-C", absPath, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
