//go:build darwin || linux

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type GoEvaluator struct {
	interp      *interp.Interpreter
	stdoutPipe  *os.File
	stderrPipe  *os.File
	originalOut *os.File
	originalErr *os.File
	state       *ShellState
	spawner     *ProcessSpawner
}

func NewGoEvaluator() *GoEvaluator {
	// Temporarily change to a clean directory to prevent auto-loading
	originalDir, _ := os.Getwd()
	tempDir := "/tmp/gosh-clean-" + fmt.Sprintf("%d", os.Getpid())
	os.MkdirAll(tempDir, 0755)
	os.Chdir(tempDir)
	
	// Create interpreter in clean directory
	i := interp.New(interp.Options{
		GoPath: os.Getenv("GOPATH"),
		Stdout: os.Stdout, // Will be updated per-eval
		Stderr: os.Stderr,
	})
	
	// Change back to original directory RIGHT AWAY (not in defer)
	os.Chdir(originalDir)
	os.RemoveAll(tempDir)

	// Load standard library
	i.Use(stdlib.Symbols)

	// Pre-import common packages for convenience
	if _, err := i.Eval(`
import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"path/filepath"
)
`); err != nil {
		fmt.Printf("Warning: Failed to preload packages: %v\n", err)
	}

	return &GoEvaluator{
		interp:      i,
		originalOut: os.Stdout,
		originalErr: os.Stderr,
	}
}

func (g *GoEvaluator) SetupWithShell(state *ShellState, spawner *ProcessSpawner) {
	g.state = state
	g.spawner = spawner
	
	// For now, we'll keep it simple and not expose shell APIs directly
	// Can extend this later with safe wrapper functions
}

func (g *GoEvaluator) stripImports(code string) string {
	lines := strings.Split(code, "\n")
	var result []string
	inImport := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "import ") {
			// Skip single-line import
			if strings.Contains(trimmed, "(") && !strings.Contains(trimmed, ")") {
				// Start of multi-line import
				inImport = true
				continue
			} else {
				// Single line import - skip it
				continue
			}
		} else if strings.HasPrefix(trimmed, "(") && !inImport {
			// Start of multi-line import block
			inImport = true
			continue
		} else if trimmed == ")" && inImport {
			// End of multi-line import block
			inImport = false
			continue
		} else if inImport {
			// Skip lines inside import block
			continue
		}
		
		result = append(result, line)
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

	// Get original directory to properly resolve relative paths
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading %s (%s): %w", configType, configPath, err)
	}

	// Strip import statements since common packages are already pre-imported
	configCode := g.stripImports(string(content))

	// Evaluate config code
	if _, err := g.interp.Eval(configCode); err != nil {
		return fmt.Errorf("error evaluating %s: %w", configType, err)
	}

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



func (g *GoEvaluator) Eval(code string) ExecutionResult {
	// Mark that we're entering yaegi evaluation
	SetYaegiEvalState(true)
	defer func() {
		SetYaegiEvalState(false)
	}()
	
	// Process command substitutions first
	processedCode := g.processCommandSubstitutions(code)

	// Check if this is a simple assignment - don't print result
	trimmed := strings.TrimSpace(processedCode)
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
				capturedOutput = formatResult(unwrapped)
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
		
		spawner := NewProcessSpawner(&ShellState{}) // Use empty state for simple command execution
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


