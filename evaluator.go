//go:build darwin || linux

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Global reference to shell state for shellapi access
var globalShellState *ShellState
var shellStateMutex sync.Mutex



type GoEvaluator struct {
	interp         *interp.Interpreter
	stdoutPipe     *os.File
	stderrPipe     *os.File
	originalOut    *os.File
	originalErr    *os.File
	state          *ShellState
	spawner        *ProcessSpawner
	builtins       *BuiltinHandler // Add builtin handler reference
	configFuncs    map[string]reflect.Value // Store config functions for calling
	
	
}

func NewGoEvaluator() *GoEvaluator {
	// Temporarily change to a clean directory to prevent auto-loading
	originalDir, _ := os.Getwd()
	tempDir := "/tmp/gosh-clean-" + fmt.Sprintf("%d", os.Getpid())
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)
	
	// Create interpreter in clean directory with unrestricted access to os/exec
	i := interp.New(interp.Options{
		GoPath:      os.Getenv("GOPATH"),
		Stdout:      os.Stdout, // Will be updated per-eval
		Stderr:      os.Stderr,
		Unrestricted: true, // Enable access to os/exec and other restricted packages
	})
	
	// Change back to original directory RIGHT AWAY (not in defer)
	os.Chdir(originalDir)
	os.RemoveAll(tempDir)

	// Load standard library
	i.Use(stdlib.Symbols)

	

	// Pre-import common packages for convenience (but NOT os/exec - will use it via shellapi functions)
	if _, err := i.Eval(`
import (
	"os"
	"strings"
	"strconv"
	"path/filepath"
)
`); err != nil {
		fmt.Printf("Warning: Failed to preload packages: %v\n", err)
	}

	
	
	
	
	// Inject shellapi functions that use os/exec internally (whitelisted via Go code)
	shellapiSymbols := map[string]map[string]reflect.Value{
		"shellapi/shellapi": {
			"RunShell": reflect.ValueOf(func(name string, args ...string) (string, error) {
				// Handle cd specially - IMMEDIATELY change the directory so it works within functions
				if name == "cd" && len(args) > 0 {
					targetPath := args[0]
					
					// Handle path expansion
					var expandedPath string
					if strings.HasPrefix(targetPath, "~") {
						home := os.Getenv("HOME")
						if len(targetPath) == 1 {
							expandedPath = home
						} else {
							expandedPath = filepath.Join(home, targetPath[1:])
						}
					} else if filepath.IsAbs(targetPath) {
						expandedPath = targetPath
					} else {
						cwd, _ := os.Getwd()
						expandedPath = filepath.Join(cwd, targetPath)
					}
					
					// Perform actual directory change immediately - THIS IS THE FIX!
					if err := os.Chdir(expandedPath); err != nil {
						return fmt.Sprintf("cd: %s: %v", targetPath, err), nil
					}
					
					// CRITICAL: Update global shell state for ALL cases (interactive and function calls)
					shellStateMutex.Lock()
					if globalShellState != nil {
						globalShellState.WorkingDirectory = expandedPath
					}
					shellStateMutex.Unlock()
					
					// Return the marker for config function calling compatibility  
					return "@GOSH_INTERNAL_CD:" + targetPath, nil
				}
				
				// Execute command using os/exec in Go code (this works - we whitelisted os/exec manually)
				cmd := exec.Command(name, args...)
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"GitStatus": reflect.ValueOf(func() (string, error) {
				cmd := exec.Command("git", "status")
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"LsColor": reflect.ValueOf(func() (string, error) {
				cmd := exec.Command("ls", "--color=auto")
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"GoBuild": reflect.ValueOf(func() (string, error) {
				cmd := exec.Command("go", "build")
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"GoTest": reflect.ValueOf(func() (string, error) {
				cmd := exec.Command("go", "test")
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"GoRun": reflect.ValueOf(func() (string, error) {
				cmd := exec.Command("go", "run", ".")
				output, err := cmd.CombinedOutput()
				return strings.TrimSpace(string(output)), err
			}),
			"Success": reflect.ValueOf(func(text string) string {
				return "\033[32m" + text + "\033[0m"
			}),
			"Warning": reflect.ValueOf(func(text string) string {
				return "\033[33m" + text + "\033[0m"
			}),
			"Error": reflect.ValueOf(func(text string) string {
				return "\033[31m" + text + "\033[0m"
			}),
		},
	}
	
	// Inject shellapi functions
	if err := i.Use(shellapiSymbols); err != nil {
		fmt.Printf("Failed to inject shellapi symbols: %v\n", err)
	}

	

	evaluator := &GoEvaluator{
		interp:      i,
		originalOut: os.Stdout,
		originalErr: os.Stderr,
		configFuncs: make(map[string]reflect.Value),
	}
	
	return evaluator
}

func (g *GoEvaluator) SetupWithShell(state *ShellState, spawner *ProcessSpawner) {
	g.state = state
	g.spawner = spawner
	
	// Set global reference for shellapi access
	shellStateMutex.Lock()
	globalShellState = state
	shellStateMutex.Unlock()
}

func (g *GoEvaluator) SetupWithBuiltins(builtins *BuiltinHandler) {
	g.builtins = builtins
}

func (g *GoEvaluator) stripImports(code string) string {
	lines := strings.Split(code, "\n")
	var result []string
	inImport := false
	shouldSkip := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		shouldSkip = false
		
		if strings.HasPrefix(trimmed, "import ") {
			if strings.Contains(line, "github.com/rsarv3006/gosh_lib/shellapi") {
				// Skip shellapi import specifically
				shouldSkip = true
			} else if strings.Contains(trimmed, "(") && !strings.Contains(trimmed, ")") {
				// Start of multi-line import - don't skip yet
				inImport = true
			} else {
				// Single line import (not shellapi) - keep it
				result = append(result, line)
				continue
			}
		} else if strings.HasPrefix(trimmed, "(") && !inImport {
			// Start of multi-line import block
			inImport = true
		} else if trimmed == ")" && inImport {
			// End of multi-line import block
			inImport = false
			result = append(result, line) // Keep the closing paren
			continue
		} else if inImport {
			// Inside multi-line import block, check if it's the shellapi import
			if strings.Contains(line, "github.com/rsarv3006/gosh_lib/shellapi") {
				shouldSkip = true
			}
		}
		
		if !shouldSkip {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

func (g *GoEvaluator) LoadConfig() error {
	// Load global config from ~/.config/gosh/config.go
	if err := g.loadConfigFile("home config", g.getHomeConfigPath()); err != nil {
		return err
	}

	return nil
}

// loadConfigFile loads a specific config file
func (g *GoEvaluator) loadConfigFile(configType, configPath string) error {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // File doesn't exist, that's OK
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading %s (%s): %w", configType, configPath, err)
	}

	

	// Define shell functions that will use command substitution
	shellCode := `
func RunShell(name string, args ...string) (string, error) {
	// Build command line
	cmdline := name
	for _, arg := range args {
		if strings.Contains(arg, " ") || strings.Contains(arg, "\"") {
			cmdline += " \"" + strings.ReplaceAll(arg, "\"", "\\\"") + "\""
		} else {
			cmdline += " " + arg
		}
	}
	
	// Return command substitution that will be processed
	return "$(" + cmdline + ")", nil
}

func ExecShell(name string, args ...string) error {
	_, err := RunShell(name, args...)
	return err
}
`

	// Evaluate shell code
	if _, err := g.interp.Eval(shellCode); err != nil {
		return fmt.Errorf("error defining shell functions: %w", err)
	}

	// Replace shellapi imports with our injected package path BEFORE stripping
	userCode := strings.ReplaceAll(string(content), `"github.com/rsarv3006/gosh_lib/shellapi"`, `"shellapi/shellapi"`)
	
	// Strip package declaration from user code (but keep all imports including shellapi/shellapi)
	lines := strings.Split(userCode, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "package main") {
			cleanLines = append(cleanLines, line)
		}
	}
	userCode = strings.Join(cleanLines, "\n")

	// Evaluate the user config code
	if _, err := g.interp.Eval(userCode); err != nil {
		return fmt.Errorf("error evaluating %s: %w", configType, err)
	}

	// Extract and store config functions for calling
	g.extractConfigFunctions()

	fmt.Printf("Loaded %s from %s\n", configType, configPath)
	return nil
}

// getHomeConfigPath returns the home config path
func (g *GoEvaluator) getHomeConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "gosh", "config.go")
}

// extractConfigFunctions finds and stores functions from the evaluated config
func (g *GoEvaluator) extractConfigFunctions() {
	// Common config functions to look for
	functionNames := []string{"gs", "build", "test", "run", "goGosh", "GitStatus", "ListFiles", "CurrentBranch", "showGo", "clean", "hello", "RunShell"}
	
	for _, funcName := range functionNames {
		// Try to evaluate the function name to get its value
		if val, err := g.interp.Eval(funcName); err == nil && val.IsValid() {
			// Store the function for later calling
			g.configFuncs[funcName] = val
		}
	}
}

// callConfigFunction attempts to call a stored config function
func (g *GoEvaluator) callConfigFunction(funcName string, args []reflect.Value) (reflect.Value, error) {
	if fn, exists := g.configFuncs[funcName]; exists {
		// Call the function with provided arguments
		if fn.Kind() == reflect.Func {
			results := fn.Call(args)
			if len(results) > 0 {
				return results[0], nil // Return first result (most common case)
			}
			return reflect.Value{}, nil // No return value
		}
	}
	return reflect.Value{}, fmt.Errorf("function %s not found", funcName)
}



func (g *GoEvaluator) Eval(code string) ExecutionResult {
	// Mark that we're entering yaegi evaluation
	SetYaegiEvalState(true)
	defer func() {
		SetYaegiEvalState(false)
	}()
	
	// Trim whitespace for checking
	trimmed := strings.TrimSpace(code)
	
	// Check if this is a bare function call from config (like "gs()" or "gs")
	// But NOT an assignment like "result := func()" 
	if funcMatch := strings.Index(trimmed, "("); funcMatch > 0 && !strings.Contains(trimmed, ":=") && !strings.Contains(trimmed, "=") {
		funcName := trimmed[:funcMatch]
		argsStr := ""
		if len(trimmed) > funcMatch+1 {
			argsStr = trimmed[funcMatch+1:]
			if argsStr[len(argsStr)-1] == ')' {
				argsStr = argsStr[:len(argsStr)-1] // Remove trailing )
			}
		}
		
		// Try to call config function
		var args []reflect.Value
		if argsStr != "" {
			// For now, only support no-argument functions like gs()
			// TODO: Parse arguments properly if needed
		}
		
		result, err := g.callConfigFunction(funcName, args)
		if err == nil {
			// Function was found and called successfully
			var output string
			if result.IsValid() {
				// Check if result contains command substitution and process it
				if result.Kind() == reflect.String {
					stringResult := result.String()
					if strings.HasPrefix(stringResult, "@GOSH_INTERNAL_CD:") {
						// Handle internal cd command
						path := strings.TrimPrefix(stringResult, "@GOSH_INTERNAL_CD:")
						if g.builtins != nil {
							cdResult := g.builtins.cd([]string{path})
							if cdResult.Error != nil {
								output = "CD ERROR: " + cdResult.Error.Error()
							} else {
								// CRITICAL: sync shell state with actual OS working directory for proper prompt display
								if currentDir, err := os.Getwd(); err == nil {
									g.state.WorkingDirectory = currentDir
								}
								output = "" // Successful cd produces no output
							}
						} else {
							output = "cd command not available in current context"
						}
					} else if strings.HasPrefix(stringResult, "$(") && strings.HasSuffix(stringResult, ")") {
						// For command substitution, process it but return RAW output, not escaped
						output = g.processCommandSubstitutionsForDisplay(stringResult)
					} else {
						output = formatResult(result)
					}
				} else {
					output = formatResult(result)
				}
			} else {
				output = ""
			}
			return ExecutionResult{
				Output:   output,
				ExitCode: 0,
				Error:    nil,
			}
		}
		// If not found in config, continue with normal evaluation
	}
	
	// Process command substitutions first
	processedCode := g.processCommandSubstitutions(code)

	// Check if this is a simple assignment - don't print result
	trimmed = strings.TrimSpace(processedCode)
	isAssignment := strings.Contains(trimmed, ":=") ||
		(strings.Contains(trimmed, "=") && !strings.Contains(trimmed, "==") &&
			!strings.Contains(trimmed, "!=") && !strings.Contains(trimmed, "<=") &&
			!strings.Contains(trimmed, ">="))

	// Check if this is a print statement - don't show return value
	isPrintStatement := strings.Contains(trimmed, "fmt.Print") ||
		strings.Contains(trimmed, "fmt.Fprint") ||
		strings.Contains(trimmed, "println(") ||
		strings.Contains(trimmed, "print(")

	// Create a pipe to capture output
	r, w, _ := os.Pipe()

	// Redirect os.Stdout and os.Stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = w
	os.Stderr = w

	// Evaluate the code with panic recovery
	var result reflect.Value
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Convert panic to error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("yaegi evaluation panic: %v", r)
				}
			}
		}()
		result, err = g.interp.Eval(processedCode)
	}()

	// Restore stdout/stderr and close write end
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	w.Close()

	// Read all captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	capturedOutput := buf.String()

	// Determine if we should show the result value
	// Show result if: no error, valid result, not an assignment, not a print, and NO stdout output
	if err == nil && result.IsValid() && !isAssignment && !isPrintStatement && len(capturedOutput) == 0 {
		// yaegi often wraps results in *interface{} - unwrap them
		unwrapped := result

		// Unwrap pointer to interface
		if result.Kind() == reflect.Ptr && result.Type().String() == "*interface {}" {
			if !result.IsNil() {
				unwrapped = result.Elem() // dereference pointer
				if unwrapped.Kind() == reflect.Interface && !unwrapped.IsNil() {
					unwrapped = unwrapped.Elem() // unwrap interface
				}
			}
		}

		// Don't print function values, invalid types, or nil values
		if unwrapped.Kind() == reflect.Func {
			// Skip function values
		} else if unwrapped.Kind() == reflect.Invalid {
			// Skip invalid values
		} else if unwrapped.Kind() == reflect.Interface && unwrapped.IsNil() {
			// Skip nil interfaces
		} else {
			// Check if it's nillable before calling IsNil
			shouldPrint := false
			if unwrapped.Kind() == reflect.Ptr || unwrapped.Kind() == reflect.Slice ||
				unwrapped.Kind() == reflect.Map || unwrapped.Kind() == reflect.Chan {
				shouldPrint = !unwrapped.IsNil()
			} else if unwrapped.Kind() == reflect.Interface {
				shouldPrint = !unwrapped.IsNil()
			} else {
				shouldPrint = true
			}

			if shouldPrint {
				formattedResult := formatResult(unwrapped)
				// Check if result contains cd marker and process it
				if strings.HasPrefix(formattedResult, "@GOSH_INTERNAL_CD:") {
					path := strings.TrimPrefix(formattedResult, "@GOSH_INTERNAL_CD:")
					if g.builtins != nil {
						cdResult := g.builtins.cd([]string{path})
						if cdResult.Error != nil {
							capturedOutput = cdResult.Output
						} else {
							capturedOutput = "" // Successful cd produces no output
						}
					} else {
						capturedOutput = "cd command not available in current context"
					}
				} else if strings.HasPrefix(formattedResult, "$(") && strings.HasSuffix(formattedResult, ")") {
					capturedOutput = g.processCommandSubstitutionsForDisplay(formattedResult)
				} else {
					capturedOutput = formattedResult
				}
			}
		}
	}

	output := strings.TrimSpace(capturedOutput)

	exitCode := 0
	if err != nil {
		exitCode = 1
		// Only add error to output if we don't already have output
		if output == "" {
			// Provide a cleaner error message for yaegi panics
			if strings.Contains(err.Error(), "CFG post-order panic") {
				output = "Go syntax error: function return type mismatch"
			} else if strings.Contains(err.Error(), "yaegi evaluation panic") {
				output = "Go syntax error: invalid Go code"
			} else {
				output = err.Error()
			}
		}
	}

	return ExecutionResult{
		Output:   strings.TrimSpace(output),
		ExitCode: exitCode,
		Error:    err,
	}
}

// EvalWithRecovery provides additional safety against yaegi crashes
func (g *GoEvaluator) EvalWithRecovery(code string) ExecutionResult {
	// Add an outer layer of recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n🚨 CRITICAL: yaegi interpreter crashed!\n")
			fmt.Fprintf(os.Stderr, "🚨 ERROR: Go evaluation may be unstable. Consider restarting.\n")
			fmt.Fprintf(os.Stderr, "🚨 ERROR: Last command was: %s\n", code[:min(len(code), 50)])
		}
	}()
	
	return g.Eval(code)
}





func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// restartInterpreter removed for simplicity - just provide crash recovery

// processCommandSubstitutionsForDisplay processes command substitutions but returns RAW output
func (g *GoEvaluator) processCommandSubstitutionsForDisplay(code string) string {
	for {
		start := strings.Index(code, "$(")
		if start == -1 {
			break
		}

		// Find matching closing parenthesis
		depth := 1
		end := -1
		for i := start + 2; i < len(code); i++ {
			if code[i] == '(' {
				depth++
			} else if code[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}

		if end == -1 {
			break // Unbalanced, return original
		}

		// Extract command
		command := code[start+2 : end]
		
		// Parse the command properly
		parts := strings.Fields(command)
		if len(parts) == 0 {
			code = code[:start] + code[end+1:] // Remove empty command
			continue
		}
		cmd := parts[0]
		args := parts[1:]
		
		spawner := NewProcessSpawner(g.state)
		result := spawner.Execute(cmd, args)
		
		// Return RAW output without any escaping
		output := result.Output

		// Replace $(command) with raw output
		code = code[:start] + output + code[end+1:]
	}

	return code
}

// processCommandSubstituions replaces $(command) with string literals containing command output
func (g *GoEvaluator) processCommandSubstitutions(code string) string {
	for {
		start := strings.Index(code, "$(")
		if start == -1 {
			break
		}

		// Find matching closing parenthesis
		depth := 1
		end := -1
		for i := start + 2; i < len(code); i++ {
			if code[i] == '(' {
				depth++
			} else if code[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}

		if end == -1 {
			break // Unbalanced, return original
		}

		// Extract command
		command := code[start+2 : end]
		
		// Execute command and get output
		// Parse the command properly
		parts := strings.Fields(command)
		if len(parts) == 0 {
			code = code[:start] + "\"\"" + code[end+1:] // Replace with empty string
			continue
		}
		cmd := parts[0]
		args := parts[1:]
		
		spawner := NewProcessSpawner(g.state) // Use current shell state for proper execution
		result := spawner.Execute(cmd, args)
		
		// Escape the output for Go string literal
		output := strings.ReplaceAll(result.Output, "\\", "\\\\")
		output = strings.ReplaceAll(output, "\"", "\\\"")
		output = strings.ReplaceAll(output, "\n", "\\n")
		output = strings.ReplaceAll(output, "\t", "\\t")
		output = strings.ReplaceAll(output, "\r", "\\r")

		// Replace $(command) with string literal
		code = code[:start] + "\"" + output + "\"" + code[end+1:]
	}

	return code
}

func formatResult(v reflect.Value) string {
	// Handle different types nicely
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// evaluateStoredConfig evaluates config content that was read before interpreter creation
func (g *GoEvaluator) evaluateStoredConfig(configType, configContent string) error {
	// Strip import statements since common packages are already pre-imported
	configContent = g.stripImports(configContent)

	// Evaluate config code
	if _, err := g.interp.Eval(configContent); err != nil {
		return fmt.Errorf("error evaluating %s: %w", configType, err)
	}

	fmt.Printf("Loaded %s\n", configType)
	return nil
}


