//go:build darwin || linux

package main

import (
	"testing"
)

func TestGoshCompleter_BasicCommandCompletion(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test that 'wh' completes to 'whoami' (suffix 'oami')
	matches, length := c.Do([]rune("wh"), 2)
	
	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}
	
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
		return
	}
	
	if string(matches[0]) != "oami" {
		t.Errorf("expected match 'oami', got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_ExactMatch(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test that exact match 'whoami' returns empty suffix
	matches, length := c.Do([]rune("whoami"), 6)
	
	if length != 6 {
		t.Errorf("expected length 6, got %d", length)
	}
	
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
		return
	}
	
	if string(matches[0]) != "" {
		t.Errorf("expected empty suffix for exact match, got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_NoMatch(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test that nonexistent command returns no matches
	matches, length := c.Do([]rune("nonexistent"), 11)
	
	if length != 11 {
		t.Errorf("expected length 11, got %d", length)
	}
	
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestGoshCompleter_MultipleMatches(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test that partial 'c' returns multiple matches
	matches, length := c.Do([]rune("c"), 1)
	
	if length != 1 {
		t.Errorf("expected length 1, got %d", length)
	}
	
	// Should have multiple matches for commands starting with 'c' (cd, cat, const)
	expectedCount := 3
	if len(matches) != expectedCount {
		t.Errorf("expected %d matches, got %d", expectedCount, len(matches))
	}
}

func TestGoshCompleter_completeCommands_Unit(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test the underlying completeCommands function directly
	matches := c.completeCommands("wh")
	
	if len(matches) != 1 {
		t.Errorf("expected 1 match for 'wh', got %d", len(matches))
		return
	}
	
	if string(matches[0]) != "oami" {
		t.Errorf("expected 'oami' suffix for 'wh', got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_completeCommands_Multiple(t *testing.T) {
	c := NewGoshCompleterForTesting()

	// Test multiple matches
	matches := c.completeCommands("c")
	
	expectedMatches := []string{"d", "at", "onst"} // cd, cat, const
	if len(matches) != len(expectedMatches) {
		t.Errorf("expected %d matches for 'c', got %d", len(expectedMatches), len(matches))
		return
	}
	
	for i, expected := range expectedMatches {
		if string(matches[i]) != expected {
			t.Errorf("expected match %d to be '%s', got '%s'", i, expected, string(matches[i]))
		}
	}
}
