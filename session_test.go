//go:build darwin || linux

package main

import (
	"strings"
	"testing"
)

func TestNewSessionState(t *testing.T) {
	s := NewSessionState()
	if s == nil {
		t.Fatal("NewSessionState returned nil")
	}
	if s.Mode != ModeShell {
		t.Errorf("Expected ModeShell, got %v", s.Mode)
	}
	if s.CapturedVars == nil {
		t.Error("CapturedVars should be initialized")
	}
	if s.History == nil {
		t.Error("History should be initialized")
	}
}

func TestSessionGetPrompt(t *testing.T) {
	s := NewSessionState()

	// Default mode is shell
	prompt := s.GetPrompt()
	if prompt != "$ " {
		t.Errorf("Expected '$ ', got %q", prompt)
	}

	// Switch to Go mode
	s.Mode = ModeGo
	prompt = s.GetPrompt()
	if prompt != "go> " {
		t.Errorf("Expected 'go> ', got %q", prompt)
	}

	// Switch back to shell
	s.Mode = ModeShell
	prompt = s.GetPrompt()
	if prompt != "$ " {
		t.Errorf("Expected '$ ', got %q", prompt)
	}
}

func TestSessionGetContinuationPrompt(t *testing.T) {
	s := NewSessionState()
	prompt := s.GetContinuationPrompt()
	if prompt != "... " {
		t.Errorf("Expected '... ', got %q", prompt)
	}
}

func TestSessionAddHistory(t *testing.T) {
	s := NewSessionState()
	s.History = nil // Reset history

	block := HistoryBlock{
		Mode:    ModeShell,
		Input:   "ls -la",
		Output:  "total 100",
		Capture: "",
	}

	s.AddHistory(block)

	if len(s.History) != 1 {
		t.Fatalf("Expected 1 history block, got %d", len(s.History))
	}

	if s.History[0].Input != "ls -la" {
		t.Errorf("Expected 'ls -la', got %q", s.History[0].Input)
	}

	// Test history limit (max 1000)
	for i := 0; i < 1005; i++ {
		s.AddHistory(HistoryBlock{
			Mode:  ModeShell,
			Input: "cmd",
		})
	}

	if len(s.History) > 1000 {
		t.Errorf("History should be limited to 1000, got %d", len(s.History))
	}
}

func TestSessionCapturedVars(t *testing.T) {
	s := NewSessionState()
	s.CapturedVars = make(map[string][]string)

	s.CapturedVars["testVar"] = []string{"line1", "line2", "line3"}

	if len(s.CapturedVars["testVar"]) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(s.CapturedVars["testVar"]))
	}
}

func TestHistoryBlockStruct(t *testing.T) {
	block := HistoryBlock{
		Mode:    ModeGo,
		Input:   "fmt.Println(\"hello\")",
		Output:  "hello",
		Capture: "output",
	}

	if block.Mode != ModeGo {
		t.Error("Mode should be ModeGo")
	}
	if !strings.Contains(block.Input, "fmt.Println") {
		t.Error("Input should contain fmt.Println")
	}
}
