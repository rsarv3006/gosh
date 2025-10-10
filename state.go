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
	cachedPrompt     string
	promptHash       string  // Content hash to detect changes
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
		CurrentProcess:   nil,
	}
}

func (s *ShellState) EnvironmentSlice() []string {
	env := make([]string, 0, len(s.Environment))
	for k, v := range s.Environment {
		env = append(env, k+"="+v)
	}
	return env
}

// GetPrompt returns a cached, colored prompt
func (s *ShellState) GetPrompt() string {
	dir := s.WorkingDirectory
	home := s.Environment["HOME"]

	if home != "" && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	gitBranch := ""
	if isInGitRepo(s.WorkingDirectory) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			gitBranch = fmt.Sprintf("git:(%s)", strings.TrimSpace(string(output)))
		}
	}

	return fmt.Sprintf("%s %s > ", dir, gitBranch)
}

// createPromptHash creates a hash of the current prompt-relevant state
func (s *ShellState) createPromptHash() string {
	hash := md5.New()
	hash.Write([]byte(s.WorkingDirectory))
	
	// Add git branch if in a git repo
	if isInGitRepo(s.WorkingDirectory) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			hash.Write(output)
		}
	}
	
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// generatePromptWithColors creates the actual colored prompt
func (s *ShellState) generatePromptWithColors() string {
	colors := GetColorManager()
	dir := s.WorkingDirectory
	home := s.Environment["HOME"]

	if home != "" && strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}

	// Style directory (only this one call to colors)
	styledDir := colors.StylePrompt(dir, "directory")

	gitBranch := ""
	if isInGitRepo(s.WorkingDirectory) {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			branchName := strings.TrimSpace(string(output))
			gitBranch = colors.StylePrompt(fmt.Sprintf("git:(%s)", branchName), "git-branch")
		}
	}

	// Style prompt symbols
	symbol := colors.StylePrompt(">", "symbol")
	space := colors.StylePrompt(" ", "separator")

	return fmt.Sprintf("%s%s%s%s%s", styledDir, space, gitBranch, space, symbol)
}

// ForcePromptRefresh can be called when we know the prompt should be updated
func (s *ShellState) ForcePromptRefresh() {
	s.promptHash = ""  // Clear hash to force refresh
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
