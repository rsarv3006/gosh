//go:build darwin || linux

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type EnvironmentManager struct {
	state *ShellState
}

func NewEnvironmentManager(state *ShellState) *EnvironmentManager {
	return &EnvironmentManager{state: state}
}

// InitializeEnvironment sets up the hybrid environment strategy
func (em *EnvironmentManager) InitializeEnvironment() {
	if em.isLoginShell() {
		em.loadLoginShellConfigs()
	} else {
		em.inheritFromParentShell()
	}

	// Ensure critical Go environment is available
	em.ensureGoEnvironment()
}

// isLoginShell checks if we're running as a login shell
func (em *EnvironmentManager) isLoginShell() bool {
	// Check if we're process 1 or have login name in argv[0]
	if len(os.Args) > 0 && strings.HasPrefix(filepath.Base(os.Args[0]), "-") {
		return true
	}

	// Check for explicit login flag
	if os.Getenv("LOGIN") == "true" {
		return true
	}

	return false
}

// loadLoginShellConfigs loads standard shell config files
func (em *EnvironmentManager) loadLoginShellConfigs() {
	home := os.Getenv("HOME")
	if home == "" {
		home = em.state.Environment["HOME"]
	}

	if home == "" {
		return
	}

	// Load system-wide configs first
	em.loadSystemConfigs()

	// Standard login config files in order of preference
	loginConfigs := []string{
		".bash_profile", // Bash login shell
		".bash_login",   // Fallback for bash
		".profile",      // POSIX standard
		".zprofile",     // Zsh login shell
		".login",        // Classic csh/sh login script
	}

	// Load login configs
	for _, configFile := range loginConfigs {
		fullPath := filepath.Join(home, configFile)
		if _, err := os.Stat(fullPath); err == nil {
			em.loadShellConfigFile(fullPath)
		}
	}

	// Also load interactive configs (for PATH and aliases)
	em.loadInteractiveConfigs(home)
}

// loadInteractiveConfigs loads interactive shell configs
func (em *EnvironmentManager) loadInteractiveConfigs(home string) {
	// Interactive configs typically contain PATH setup and aliases
	interactiveConfigs := []string{
		".zshrc",        // Zsh interactive
		".bashrc",       // Bash interactive
		".bash_aliases", // Bash aliases
	}

	for _, configFile := range interactiveConfigs {
		fullPath := filepath.Join(home, configFile)
		if _, err := os.Stat(fullPath); err == nil {
			em.loadShellConfigFile(fullPath)
		}
	}
}

// loadSystemConfigs loads system-wide shell configurations
func (em *EnvironmentManager) loadSystemConfigs() {
	systemConfigs := []string{
		"/etc/zprofile",   // Zsh system-wide login
		"/etc/profile",    // POSIX system-wide login
		"/etc/bashrc",     // Bash system-wide interactive
		"/etc/zshrc",      // Zsh system-wide interactive
	}

	for _, configPath := range systemConfigs {
		if _, err := os.Stat(configPath); err == nil {
			em.loadShellConfigFile(configPath)
		}
	}
}

// loadShellConfigFile loads and executes a standard shell config file
func (em *EnvironmentManager) loadShellConfigFile(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse basic shell export statements
		if strings.HasPrefix(line, "export ") {
			em.parseExport(line[7:]) // Remove "export " prefix
		} else if strings.Contains(line, "=") && !strings.ContainsAny(line, "(){}") {
			// Simple variable assignment
			em.parseExport(line)
		}
	}
}

// parseExport handles export VAR=value and VAR=value syntax
func (em *EnvironmentManager) parseExport(exportLine string) {
	parts := strings.SplitN(exportLine, "=", 2)
	if len(parts) != 2 {
		return
	}

	varName := strings.TrimSpace(parts[0])
	varValue := strings.TrimSpace(parts[1])

	// Remove quotes if present
	if strings.HasPrefix(varValue, `"`) && strings.HasSuffix(varValue, `"`) {
		varValue = varValue[1 : len(varValue)-1]
	} else if strings.HasPrefix(varValue, `'`) && strings.HasSuffix(varValue, `'`) {
		varValue = varValue[1 : len(varValue)-1]
	}

	// Handle $HOME and other variable substitutions
	varValue = em.expandVariables(varValue)

	// Set in our environment
	em.state.Environment[varName] = varValue
}

// expandVariables expands shell variables like $HOME, $USER, etc.
func (em *EnvironmentManager) expandVariables(input string) string {
	result := input
	// Expand common variables
	home := em.state.Environment["HOME"]
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home != "" {
		result = strings.ReplaceAll(result, "$HOME", home)
		result = strings.ReplaceAll(result, "~", home)
	}

	// Expand $USER
	user := em.state.Environment["USER"]
	if user == "" {
		user = os.Getenv("USER")
	}
	if user != "" {
		result = strings.ReplaceAll(result, "$USER", user)
	}

	// Simple PATH expansion (could be enhanced)
	if strings.Contains(result, "$PATH") {
		path := em.state.Environment["PATH"]
		if path == "" {
			path = os.Getenv("PATH")
		}
		result = strings.ReplaceAll(result, "$PATH", path)
	}

	return result
}

// inheritFromParentShell cleans up environment inheritance
func (em *EnvironmentManager) inheritFromParentShell() {
	// Environment is already captured in NewShellState()
	// Try to load interactive configs to get PATH and aliases
	home := os.Getenv("HOME")
	if home == "" {
		home = em.state.Environment["HOME"]
	}
	if home != "" {
		em.loadInteractiveConfigs(home)
	}
	
	// Ensure critical variables are present
	em.ensureCriticalEnvVars()
}

// ensureCriticalEnvVars ensures essential environment variables are set
func (em *EnvironmentManager) ensureCriticalEnvVars() {
	criticalVars := map[string]string{
		"HOME":     "/Users/" + os.Getenv("USER"),
		"USER":     os.Getenv("USER"),
		"LANG":     "en_US.UTF-8",
		"TERM":     "xterm-256color",
		"SHELL":    "/bin/bash", // Fallback
	}

	// NOTE: We DON'T set default PATH here to avoid overriding inherited PATH

	for varName, defaultValue := range criticalVars {
		if em.state.Environment[varName] == "" {
			if existing := os.Getenv(varName); existing != "" {
				em.state.Environment[varName] = existing
			} else {
				em.state.Environment[varName] = defaultValue
			}
		}
	}

	// Set SHELL to our path if we can determine it
	if exePath, err := os.Executable(); err == nil {
		em.state.Environment["SHELL"] = exePath
	}
}

// ensureGoEnvironment sets up Go-specific environment
func (em *EnvironmentManager) ensureGoEnvironment() {
	// WORKAROUND: Add missing local bin directory if not present
	home := em.state.Environment["HOME"]
	if home != "" {
		localBin := filepath.Join(home, ".local", "bin")
		path := em.state.Environment["PATH"]
		if !strings.Contains(path, localBin) {
			em.state.Environment["PATH"] = localBin + ":" + path
		}
	}

	// GOPATH
	if em.state.Environment["GOPATH"] == "" {
		if gopath := os.Getenv("GOPATH"); gopath != "" {
			em.state.Environment["GOPATH"] = gopath
		} else {
			// Default GOPATH
			home := em.state.Environment["HOME"]
			if home != "" {
				em.state.Environment["GOPATH"] = filepath.Join(home, "go")
			}
		}
	}

	// GOMODCACHE
	if em.state.Environment["GOMODCACHE"] == "" {
		if gomodcache := os.Getenv("GOMODCACHE"); gomodcache != "" {
			em.state.Environment["GOMODCACHE"] = gomodcache
		} else if gopath := em.state.Environment["GOPATH"]; gopath != "" {
			em.state.Environment["GOMODCACHE"] = filepath.Join(gopath, "pkg", "mod")
		}
	}

	// Go-related PATH additions
	goBin := filepath.Join(em.state.Environment["GOPATH"], "bin")
	if goBin != "" {
		path := em.state.Environment["PATH"]
		if !strings.Contains(path, goBin) {
			em.state.Environment["PATH"] = goBin + ":" + path
		}
	}
}
