//go:build darwin || linux

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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
}

func NewGoEvaluator() *GoEvaluator {
	i := interp.New(interp.Options{
		GoPath: os.Getenv("GOPATH"),
		Stdout: os.Stdout, // Will be updated per-eval
		Stderr: os.Stderr,
	})

	// Load standard library
	i.Use(stdlib.Symbols)

	// Pre-import common packages for convenience
	i.Eval(`
import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"path/filepath"
)
`)

	return &GoEvaluator{
		interp:      i,
		originalOut: os.Stdout,
		originalErr: os.Stderr,
	}
}

func (g *GoEvaluator) Eval(code string) ExecutionResult {
	// Check if this is a simple assignment - don't print result
	trimmed := strings.TrimSpace(code)
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

	// Evaluate the code
	result, err := g.interp.Eval(code)

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
		// Check if it's nillable before calling IsNil
		shouldPrint := false
		if result.Kind() == reflect.Ptr || result.Kind() == reflect.Interface ||
			result.Kind() == reflect.Slice || result.Kind() == reflect.Map ||
			result.Kind() == reflect.Chan || result.Kind() == reflect.Func {
			shouldPrint = !result.IsNil()
		} else {
			shouldPrint = true
		}

		if shouldPrint {
			capturedOutput = formatResult(result)
		}
	}

	output := strings.TrimSpace(capturedOutput)

	exitCode := 0
	if err != nil {
		exitCode = 1
		// Only add error to output if we don't already have output
		if output == "" {
			output = err.Error()
		}
	}

	return ExecutionResult{
		Output:   strings.TrimSpace(output),
		ExitCode: exitCode,
		Error:    err,
	}
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
