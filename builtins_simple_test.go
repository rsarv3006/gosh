//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNewBuiltinHandler(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	if handler == nil {
		t.Fatal("NewBuiltinHandler returned nil")
	}

	if handler.state != state {
		t.Error("handler.state not set correctly")
	}
}

func TestBuiltinHandler_IsBuiltin(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	tests := []struct {
		command  string
		expected bool
	}{
		{"cd", true},
		{"exit", true},
		{"help", true},
		{"ls", false},
		{"git", false},
		{"echo", false},
		{"", false},
		{"CD", false}, // case sensitive
		{"cd2", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := handler.IsBuiltin(tt.command)
			if result != tt.expected {
				t.Errorf("IsBuiltin(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestBuiltinHandler_Exit_Basic(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	result := handler.exit([]string{})

	if result.ExitCode != 0 {
		t.Errorf("Exit code should be 0 for exit(), got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Error should be nil for successful exit, got %v", result.Error)
	}

	if result.Output != "" {
		t.Errorf("Output should be empty for exit(), got %q", result.Output)
	}

	if !state.ShouldExit {
		t.Error("ShouldExit should be true after exit() call")
	}
}

func TestBuiltinHandler_Exit_WithCode(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	codes := []int{0, 1, 127, 255}

	for _, code := range codes {
		t.Run(fmt.Sprintf("code_%d", code), func(t *testing.T) {
			// Reset state
			state.ShouldExit = false
			state.ExitCode = 0

			result := handler.exit([]string{fmt.Sprintf("%d", code)})

			if result.ExitCode != 0 {
				t.Errorf("Exit code should be 0 for exit() result, got %d", result.ExitCode)
			}

			if state.ExitCode != code {
				t.Errorf("State.ExitCode should be %d, got %d", code, state.ExitCode)
			}

			if !state.ShouldExit {
				t.Error("ShouldExit should be true after exit() call")
			}
		})
	}
}

func TestBuiltinHandler_CD_Success(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Original working directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	state := &ShellState{WorkingDirectory: originalDir}
	handler := NewBuiltinHandler(state)

	result := handler.cd([]string{tempDir})

	if result.ExitCode != 0 {
		t.Errorf("Exit code should be 0 for successful cd, got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Error should be nil for successful cd, got %v", result.Error)
	}

	if result.Output != "" {
		t.Errorf("Output should be empty for successful cd, got %q", result.Output)
	}

	if state.WorkingDirectory != tempDir {
		t.Errorf("Working directory should be %q, got %q", tempDir, state.WorkingDirectory)
	}
}

func TestBuiltinHandler_CD_Error(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	// Try to cd to a non-existent directory
	nonExistentPath := "/path/that/does/not/exist/12345"
	result := handler.cd([]string{nonExistentPath})

	if result.ExitCode == 0 {
		t.Errorf("Exit code should be non-zero for failed cd")
	}

	if result.Error == nil {
		t.Error("Error should not be nil for failed cd")
	}

	expectedPrefix := "cd: " + nonExistentPath + ": "
	if result.Output == "" {
		t.Error("Output should contain error message")
	} else if !strings.HasPrefix(result.Output, expectedPrefix) {
		t.Errorf("Output should start with %q, got %q", expectedPrefix, result.Output)
	}
}

func TestBuiltinHandler_Help_General(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	result := handler.help([]string{})

	if result.ExitCode != 0 {
		t.Errorf("Exit code should be 0 for help, got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Error should be nil for successful help, got %v", result.Error)
	}

	// Check that help output contains expected sections
	expectedSections := []string{
		"gosh - Go Shell with yaegi interpreter",
		"COMMANDS:",
		"cd [DIR]",
		"exit [CODE]",
		"help [COMMAND]",
		"GO CODE:",
		"COMMAND SUBSTITUTION:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result.Output, section) {
			t.Errorf("Help output should contain %q", section)
		}
	}
}

func TestBuiltinHandler_Help_Command(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	commands := []string{"cd", "exit", "help"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			result := handler.help([]string{cmd})

			if result.ExitCode != 0 {
				t.Errorf("Exit code should be 0 for help %s, got %d", cmd, result.ExitCode)
			}

			if result.Error != nil {
				t.Errorf("Error should be nil for successful help %s, got %v", cmd, result.Error)
			}

			if !strings.Contains(result.Output, cmd) {
				t.Errorf("Help output for %s should contain command name", cmd)
			}

			if !strings.Contains(result.Output, "USAGE:") {
				t.Errorf("Help output for %s should contain USAGE section", cmd)
			}
		})
	}
}

func TestBuiltinHandler_Help_Error(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	// Test help for non-existent command
	result := handler.help([]string{"nonexistentcommand123"})

	if result.ExitCode == 0 {
		t.Error("Exit code should be non-zero for help on non-existent command")
	}

	if result.Error == nil {
		t.Error("Error should not be nil for help on non-existent command")
	}

	expected := "No help available for 'nonexistentcommand123'"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

func TestBuiltinHandler_Execute_UnkownCommand(t *testing.T) {
	state := NewShellState()
	handler := NewBuiltinHandler(state)

	result := handler.Execute("unknowncommand", []string{})

	if result.ExitCode == 0 {
		t.Error("Exit code should be non-zero for unknown builtin")
	}

	if result.Error == nil {
		t.Error("Error should not be nil for unknown builtin")
	}

	expected := "Unknown builtin: unknowncommand"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

func TestBuiltinHandler_ExpandPath(t *testing.T) {
	// Create a temporary directory to use as fake home
	tempDir, err := os.MkdirTemp("", "gosh-test-home")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// This is indirect testing through cd functionality
	state := &ShellState{
		WorkingDirectory: "/some/dir",
		Environment: map[string]string{
			"HOME": tempDir,
		},
	}
	handler := NewBuiltinHandler(state)

	// Test cd with ~ expansion
	result := handler.cd([]string{"~"})
	if result.ExitCode != 0 {
		t.Errorf("cd ~ should work, got error: %v", result.Error)
	}

	if state.WorkingDirectory != tempDir {
		t.Errorf("cd ~ should expand to %s, got %q", tempDir, state.WorkingDirectory)
	}
}
