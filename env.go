//go:build darwin || linux

package main

import (
	"bufio"
	"os"
	"os/exec"
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
	isLogin := em.isLoginShell()
	
	if isLogin {
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

		// Handle eval statements with commands
		if strings.HasPrefix(line, "eval ") {
			em.handleEvalStatement(line[5:]) // Remove "eval " prefix
		} else if strings.HasPrefix(line, "export ") {
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

// handleEvalStatement processes eval commands, especially for brew shellenv
func (em *EnvironmentManager) handleEvalStatement(evalLine string) {
	// Handle brew shellenv specifically
	if strings.Contains(evalLine, "brew shellenv") {
		em.handleBrewShellenv()
		return
	}
	
	// TODO: Handle other eval statements if needed
}

// handleBrewShellenv executes brew shellenv and captures its output
func (em *EnvironmentManager) handleBrewShellenv() {
	// Try common brew locations
	brewPaths := []string{
		"/opt/homebrew/bin/brew", // Apple Silicon
		"/usr/local/bin/brew",     // Intel
		"/home/linuxbrew/.linuxbrew/bin/brew", // Linux
	}
	
	var brewPath string
	for _, path := range brewPaths {
		if _, err := os.Stat(path); err == nil {
			brewPath = path
			break
		}
	}
	
	if brewPath == "" {
		return
	}

	// Execute brew shellenv to get the actual environment
	cmd := exec.Command(brewPath, "shellenv")
	cmd.Env = em.getAllEnvVars()
	output, err := cmd.Output()
	if err != nil {
		// Simple fallback: manually add brew paths
		em.addBrewPaths()
		return
	}

	// Execute the shellenv commands in a subshell to get the final environment
	shellCommands := string(output) + " && env"
	shellCmd := exec.Command("/bin/bash", "-c", shellCommands)
	shellCmd.Env = em.getAllEnvVars()
	envOutput, err := shellCmd.Output()
	if err != nil {
		// Fallback: manually add brew paths
		em.addBrewPaths()
		return
	}

	// Parse the environment output
	em.parseEnvOutput(string(envOutput))
}

// addBrewPaths is fallback for when shellenv execution fails
func (em *EnvironmentManager) addBrewPaths() {
	currentPath := em.state.Environment["PATH"]
	brewBin := "/opt/homebrew/bin"
	brewSBin := "/opt/homebrew/sbin"
	
	if !strings.Contains(currentPath, brewBin) {
		if _, err := os.Stat(brewBin); err == nil {
			em.state.Environment["PATH"] = brewBin + ":" + currentPath
		}
	}
	
	if !strings.Contains(em.state.Environment["PATH"], brewSBin) {
		if _, err := os.Stat(brewSBin); err == nil {
			em.state.Environment["PATH"] = em.state.Environment["PATH"] + ":" + brewSBin
		}
	}
}

// getAllEnvVars gets current environment as slice
func (em *EnvironmentManager) getAllEnvVars() []string {
	env := make([]string, 0, len(em.state.Environment))
	for k, v := range em.state.Environment {
		env = append(env, k+"="+v)
	}
	
	// Add current process env as fallback
	for _, v := range os.Environ() {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			if _, exists := em.state.Environment[parts[0]]; !exists {
				env = append(env, v)
			}
		}
	}
	
	return env
}

// parseEnvOutput parses env - format output (KEY=value\nKEY=value)
func (em *EnvironmentManager) parseEnvOutput(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				em.state.Environment[parts[0]] = parts[1]
			}
		}
	}
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
