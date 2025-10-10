# Gosh Regression Testing Guide

This document provides comprehensive manual testing steps to ensure gosh continues working correctly after changes.

## Quick Regression Checklist

Run these as a baseline test after any significant changes:

### 1. Basic Go Evaluation (5 minutes)

```bash
echo -e "x := 42\nx\nexit" | ./gosh
# Expected output: 42

echo -e "name := \"gosh\"\nname\nexit" | ./gosh  
# Expected output: gosh

echo -e "a, b := 5, 3\na + b\nexit" | ./gosh
# Expected output: 8
```

### 2. Function Definitions (3 minutes)

```bash
echo -e "func add(a int, b int) int { return a + b }\nadd(2, 3)\nexit" | ./gosh
# Expected output: 5

echo -e "func greet(name string) { fmt.Printf(\"Hello %s\\n\", name) }\ngreet(\"test\")\nexit" | ./gosh
# Expected output: Hello test
```

### 3. Command Substitution (2 minutes)

```bash
echo -e "files := \$(ls)\nlen(strings.Split(files, \"\\n\"))\nexit" | ./gosh
# Expected: A number > 0 showing file count

echo -e "user := \$(whoami)\nuser\nexit" | ./gosh
# Expected: Your username
```

### 4. Shell Commands (2 minutes)

```bash
echo -e "echo 'shell works'\nwhoami\npwd\nexit" | ./gosh
# Expected: echo output, username, current directory

echo -e "cd /tmp && pwd\nexit" | ./gosh
# Expected: /tmp
```

### 5. Built-in Commands (2 minutes)

```bash
echo -e "pwd\nhelp\nexit" | ./gosh
# Expected: Current dir, then help text

echo -e "cd /nonexistent\nexit" | ./gosh
# Expected: Error message
```

### 6. Multiline Support (1 minute)

```bash
echo -e "for i := 0; i < 3; i++ {\nfmt.Println(i)\n}\nexit" | ./gosh
# Expected:
# 0
# 1
# 2
```

### 7. Variable Persistence (1 minute)

```bash
echo -e "counter := 0\ncounter = counter + 1\ncounter\ncounter = counter + 1\ncounter\nexit" | ./gosh
# Expected:
# 1
# 2
```

## Comprehensive Regression Testing

These should be run before releases or major changes.

### Core Language Features

#### Assignment and Variables
```bash
# Basic assignment
echo -e "x := 42\nx\nexit" | ./gosh

# Multiple assignment  
echo -e "a, b := 1, 2\na\nb\nexit" | ./gosh

# String assignment
echo -e "s := \"hello\"\ns\nexit" | ./gosh

# Reassignment
echo -e "x := 10\nx = 20\nx\nexit" | ./gosh
```

#### Type Declarations
```bash
# Variable declaration
echo -e "var num int = 100\nnum\nexit" | ./gosh

# Constants
echo -e "const pi = 3.14\npi\nexit" | ./gosh

# Arrays
echo -e "arr := [3]int{1, 2, 3}\narr[0]\nexit" | ./gosh

# Slices
echo -e "slice := []int{4, 5, 6}\nslice[1]\nexit" | ./gosh
```

#### Control Flow
```bash
# If statements
echo -e "x := 10\nif x > 5 {\nfmt.Println(\"big\")\n} else {\nfmt.Println(\"small\")\n}\nexit" | ./gosh

# For loops
echo -e "sum := 0\nfor i := 1; i <= 3; i++ {\nsum += i\n}\nsum\nexit" | ./gosh

# Range loops
echo -e "nums := []int{10, 20}\ntotal := 0\nfor _, n := range nums {\ntotal += n\n}\ntotal\nexit" | ./gosh
```

#### Functions
```bash
# No parameters, no return
echo -e "func hello() { fmt.Println(\"hi\") }\nhello()\nexit" | ./gosh

# With parameters
echo -e "func double(x int) int { return x * 2 }\ndouble(5)\nexit" | ./gosh

# Multiple returns
echo -e "func divmod(a, b int) (int, int) { return a / b, a % b }\ndivmod(7, 3)\nexit" | ./gosh

# Closures
echo -e "adder := func(x int) func(int) int { return func(y int) int { return x + y } }\nadd5 := adder(5)\nadd5(3)\nexit" | ./gosh
```

#### Built-in Types
```bash
# Maps
echo -e "m := map[string]int{\"a\": 1, \"b\": 2}\nm[\"a\"]\nexit" | ./gosh

# Structs  
echo -e "type Point struct { X, Y int }\np := Point{1, 2}\np.X\nexit" | ./gosh

# Interfaces
echo -e "var w io.Writer\nw = os.Stdout\nfmt.Fprintf(w, \"test\\n\")\nexit" | ./gosh
```

### Shell Integration

#### Command Execution
```bash
# Simple commands
echo -e "echo test\nwhoami\npwd\nls\nexit" | ./gosh

# Commands with flags
echo -e "ls -la\ngrep -n 'func' *.go\nexit" | ./gosh

# Commands with args
echo -e "echo hello world\ntouch testfile && ls testfile\nexit" | ./gosh
```

#### Command Substitution
```bash
# Basic substitution
echo -e "output := \$(echo hello)\noutput\nexit" | ./gosh

# With flags
echo -e "files := \$(ls -1 | head -2)\nfiles\nexit" | ./gosh

# In expressions
echo -e "count := len(strings.Split(\$(ls), \"\\n\"))\ncount\nexit" | ./gosh
```

#### Built-in Commands
```bash
# cd command
echo -e "pwd\ncd /tmp\npwd\ncd -\npwd\nexit" | ./gosh

# pwd command  
echo -e "pwd -L\npwd\nexit" | ./gosh

# exit command
echo -e "exit 42; echo 'should not see this'" | ./gosh
# Should exit with code 42

# help command
echo -e "help\nhelp cd\nhelp go\nexit" | ./gosh
```

### Error Handling

#### Syntax Errors
```bash
# Missing brace
echo -e "if true {\nfmt.Println(\"test\")\nexit" | ./gosh
# Expected: Syntax error

# Type mismatch
echo -e "var s string = 42\ns\nexit" | ./gosh
# Expected: Type error

# Undefined variable
echo -e "undefined_var\nexit" | ./gosh
# Expected: Undefined error
```

#### Command Errors
```bash
# Nonexistent command
echo -e "nonexistent_cmd\nexit" | ./gosh
# Expected: Command not found

# Command failure
echo -e "false; echo 'should see this'\ntrue; echo 'success'\nexit" | ./gosh

# Bad directory
echo -e "cd /nonexistent/directory\nexit" | ./gosh
# Expected: Error message
```

### Edge Cases

#### Multiline Input
```bash
# Complex multiline
echo -e "func factorial(n int) int {\nif n <= 1 {\nreturn 1\n}\nreturn n * factorial(n-1)\n}\nfactorial(5)\nexit" | ./gosh

# Incomplete multiline detection
echo -e "if true {\nfmt.Println(\"line 1\\nline 2\")\n}\nexit" | ./gosh
```

#### State Persistence
```bash
# Variables persist across commands
echo -e "x := 100\nfmt.Println(x)\nx = 200\nfmt.Println(x)\nexit" | ./gosh

# Functions persist
echo -e "func test() { return 42 }\ntest()\ntest()\nexit" | ./gosh
```

#### Special Characters
```bash
# Quotes and escapes
echo -e "s := \"hello \\\"world\\\"\"\ns\nexit" | ./gosh

# Paths with spaces
echo -e "path := \$(pwd)\npath\nexit" | ./gosh

# Unicode
echo -e "msg := \"测试\"\nmsg\nexit" | ./gosh
```

## Configuration Testing

Test config.go loading:

```bash
# Create test config
cat > test_config.go << 'EOF'
package main
import "fmt"
func init() {
    fmt.Println("Config loaded")
    func custom() string { return "from config" }
    custom_var := 123
}
EOF

echo -e "custom()\ncustom_var\nexit" | ./gosh
# Expected: "Config loaded", "from config", 123

rm test_config.go
```

## Performance Testing

### Large Outputs
```bash
# Large command substitution
echo -e "long_output := \$(seq 1 1000)\nlen(strings.Split(long_output, \"\\n\"))\nexit" | ./gosh

# Large Go arrays
echo -e "arr := make([]int, 10000)\nfor i := range arr { arr[i] = i }\narr[9999]\nexit" | ./gosh
```

### Memory Usage
```bash
# Monitor memory during heavy usage
echo -e "for i := 0; i < 100; i++ {\nfmt.Printf(\"Iteration %d\\n\", i)\n}\nexit" | ./gosh
# Watch memory usage with Activity Monitor/htop
```

## Automated Testing Commands

Create a test script to run all regression tests:

```bash
#!/bin/bash
# regression_test.sh

set -e
GOSH_BINARY="./gosh"

echo "Starting gosh regression tests..."

# Test 1: Basic Go evaluation
echo "Test 1: Basic Go evaluation"
echo -e "x := 42\nx\nexit" | $GOSH_BINARY | grep -q "42" || { echo "FAILED: Basic Go evaluation"; exit 1; }
echo "✓ Basic Go evaluation"

# Test 2: Functions
echo "Test 2: Function definitions"
echo -e "func add(a, b int) int { return a + b }\nadd(2, 3)\nexit" | $GOSH_BINARY | grep -q "5" || { echo "FAILED: Function definitions"; exit 1; }
echo "✓ Function definitions"

# Test 3: Command substitution
echo "Test 3: Command substitution"
echo -e "user := \$(whoami)\nuser\nexit" | $GOSH_BINARY | grep -q "$(whoami)" || { echo "FAILED: Command substitution"; exit 1; }
echo "✓ Command substitution"

# Test 4: Shell commands
echo "Test 4: Shell commands"
echo -e "echo test\nexit" | $GOSH_BINARY | grep -q "test" || { echo "FAILED: Shell commands"; exit 1; }
echo "✓ Shell commands"

# Test 5: Builtins
echo "Test 5: Built-in commands"
echo -e "pwd\nexit" | $GOSH_BINARY | grep -q "/Users" && echo "✓ Built-in commands" || { echo "FAILED: Built-in commands"; exit 1; }

echo "All regression tests passed! ✓"
```

## Environment Testing

Test in different environments:

### Terminal Capabilities
```bash
# Test in color terminal
./gosh < test_input.txt

# Test with NO_COLOR
NO_COLOR=1 ./gosh < test_input.txt  

# Test in non-interactive pipe
echo "test" | ./gosh -c "x := 42; x"
```

### Different Shells
```bash
# Test if gosh works properly when launched from different shells
bash -c "./gosh -c 'x := 42; x'"
zsh -c "./gosh -c 'x := 42; x'"
fish -c "./gosh -c 'x := 42; x'"
```

## Troubleshooting Common Issues

### "command not found" errors
- Check if the input looks like Go code but is being routed to shell
- Verify router logic in `router.go`
- Test with the exact input that fails

### Yaegi panics/errors
- Check Go syntax is valid
- Test simpler versions of the failing code
- Check for missing imports or type issues

### Configuration loading issues
- Verify config.go syntax
- Check file permissions
- Test with empty config

## Release Checklist

Before releasing, run:

1. ✅ Quick Regression Checklist (all 7 sections)
2. ✅ Comprehensive Core Language Features
3. ✅ Shell Integration tests  
4. ✅ Error Handling tests
5. ✅ Edge Cases
6. ✅ Configuration Testing
7. ✅ Performance Testing (basic)
8. ✅ Environment Testing (NO_COLOR, pipes)
9. ✅ Run existing unit tests
10. ✅ Manual interactive test session (15 minutes)

## Reporting Issues

When reporting regression issues, include:

1. Exact commands that failed
2. Expected vs actual output  
3. gosh version and go version
4. Operating system
5. Whether issue occurs in fresh checkout
6. Any error messages or panics

For example:
```
Issue: Function definition with parameters fails
Command: echo -e "func add(a, b int) int { return a + b }\nadd(2, 3)\nexit" | ./gosh
Expected: 5
Actual: Azure Functions help text
OS: macOS 14.0
Go version: 1.21.0
```
