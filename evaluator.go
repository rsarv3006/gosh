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
	interp *interp.Interpreter
}

func NewGoEvaluator() *GoEvaluator {
	i := interp.New(interp.Options{
		GoPath: os.Getenv("GOPATH"),
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

	return &GoEvaluator{interp: i}
}

func (g *GoEvaluator) Eval(code string) ExecutionResult {
	// Check if this is a simple assignment - don't print result
	isAssignment := strings.Contains(strings.TrimSpace(code), ":=") ||
		(strings.Contains(code, "=") && !strings.Contains(code, "==") &&
			!strings.Contains(code, "!=") && !strings.Contains(code, "<=") &&
			!strings.Contains(code, ">="))

	// Capture stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Evaluate the code
	result, err := g.interp.Eval(code)

	// Restore stdout/stderr
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// If there's a result value and no explicit output, print it
	// But skip assignments and if we already printed to stdout
	if err == nil && result.IsValid() && !isAssignment {
		// Check if it's nillable before calling IsNil
		shouldPrint := false
		if result.Kind() == reflect.Ptr || result.Kind() == reflect.Interface ||
			result.Kind() == reflect.Slice || result.Kind() == reflect.Map ||
			result.Kind() == reflect.Chan || result.Kind() == reflect.Func {
			shouldPrint = !result.IsNil()
		} else {
			shouldPrint = true
		}

		// Only show result if we didn't already print output
		if shouldPrint && output == "" {
			output = formatResult(result)
		}
	}

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
