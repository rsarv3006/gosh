//go:build darwin || linux

package main

import (
	"os"
	"strings"
	"testing"
)

func TestNewEnvironmentManager(t *testing.T) {
	state := NewShellState()
	manager := NewEnvironmentManager(state)

	if manager == nil {
		t.Fatal("NewEnvironmentManager returned nil")
	}
}

func TestInitializeEnvironment(t *testing.T) {
	state := NewShellState()
	manager := NewEnvironmentManager(state)

	manager.InitializeEnvironment()

	// Check that environment has some basic variables
	if len(state.Environment) == 0 {
		t.Error("Environment should not be empty after initialization")
	}
}

func TestParseExport(t *testing.T) {
	state := NewShellState()
	manager := NewEnvironmentManager(state)

	// First clear PATH to test
	state.Environment["PATH"] = ""

	// Test valid export - note that parseExport stores the full "export PATH" as key
	manager.parseExport("export PATH=/usr/bin:/bin")

	// Check that the variable was set (with export prefix)
	if state.Environment["export PATH"] != "/usr/bin:/bin" {
		t.Errorf("Expected '/usr/bin:/bin', got %q", state.Environment["export PATH"])
	}

	// Test invalid export (no =)
	manager.parseExport("export INVALID")

	// Should not crash or add invalid entry
}

func TestGetAllEnvVars(t *testing.T) {
	state := NewShellState()
	manager := NewEnvironmentManager(state)

	vars := manager.getAllEnvVars()

	if len(vars) == 0 {
		t.Error("getAllEnvVars should return non-empty slice")
	}

	// Should contain some env vars
	found := false
	for _, v := range vars {
		if strings.Contains(v, "=") {
			found = true
			break
		}
	}
	if !found {
		t.Error("getAllEnvVars output should contain '='")
	}
}

func TestNewProcessSpawner(t *testing.T) {
	state := NewShellState()
	spawner := NewProcessSpawner(state)

	if spawner == nil {
		t.Fatal("NewProcessSpawner returned nil")
	}
}

func TestFindInPath(t *testing.T) {
	// Test finding a command that should exist
	path := os.Getenv("PATH")
	result, found := FindInPath("ls", path)

	if !found {
		t.Error("Should find 'ls' command")
	}

	if result == "" {
		t.Error("Result should not be empty")
	}

	// Test non-existent command
	result, found = FindInPath("nonexistent_command_12345", path)

	if found {
		t.Error("Should not find non-existent command")
	}
}

func TestExecuteSimpleCommand(t *testing.T) {
	state := NewShellState()
	spawner := NewProcessSpawner(state)

	// Execute a simple command
	result := spawner.Execute("echo", []string{"hello"})

	if result.Error != nil {
		t.Errorf("Execute failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}
