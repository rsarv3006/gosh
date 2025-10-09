//go:build darwin || linux

package main

import (
	"testing"
)

func TestRouter_Debug_CommandRouting(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	testCases := []string{
		"ls",
		"ls -la", 
		"git status",
		"echo hello",
	}

	for _, input := range testCases {
		t.Run("Debug_"+input, func(t *testing.T) {
			inputType, command, args := router.Route(input)
			
			t.Logf("Input: %q", input)
			t.Logf("InputType: %v", inputType)
			t.Logf("Command: %q", command)
			t.Logf("Args: %v", args)
			t.Log("---")
		})
	}
}

func TestRouter_LooksLikeShellCommand_Debug(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	testCases := []string{
		"ls",
		"ls -la",
		"git status",
		"echo hello world",
	}

	for _, input := range testCases {
		t.Run("DebugLooksLikeShellCommand_"+input, func(t *testing.T) {
			result := router.looksLikeShellCommand(input)
			t.Logf("Input: %q", input)
			t.Logf("LooksLikeShellCommand: %v", result)
			t.Log("---")
		})
	}
}

func TestRouter_ParseInput_Debug(t *testing.T) {
	builtins := NewBuiltinHandler(NewShellState())
	router := NewRouter(builtins)

	testCases := []string{
		"ls",
		"ls -la",
		"git status",
		"echo hello world",
	}

	for _, input := range testCases {
		t.Run("DebugParseInput_"+input, func(t *testing.T) {
			command, args := router.parseInput(input)
			t.Logf("Input: %q", input)
			t.Logf("Command: %q", command)
			t.Logf("Args: %v", args)
			t.Log("---")
		})
	}
}
