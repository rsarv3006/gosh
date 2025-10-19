//go:build darwin || linux

package main

import (
	"testing"
)

func TestGoshCompleter_BasicCommandCompletion(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

	// Test that 'wh' finds matches (could be whoami or other PATH executables)
	matches, length := c.Do([]rune("wh"), 2)
	
	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}
	
	if len(matches) < 1 {
		t.Errorf("expected at least 1 match, got %d", len(matches))
		return
	}
	
	// Should find something starting with 'wh'
	if len(matches[0]) < 1 {
		t.Errorf("expected non-empty completion suffix, got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_ExactMatch(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

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
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

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
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

	// Test that partial 'c' returns multiple matches
	matches, length := c.Do([]rune("c"), 1)
	
	if length != 1 {
		t.Errorf("expected length 1, got %d", length)
	}
	
	// Should have multiple matches for commands starting with 'c' (cd, cat, plus PATH executables)
	expectedMinCount := 2 // At least cd, cat 
	if len(matches) < expectedMinCount {
		t.Errorf("expected at least %d matches, got %d", expectedMinCount, len(matches))
	}
}

func TestGoshCompleter_completeCommands_Unit(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

	// Test the underlying completeCommands function directly
	matches := c.completeCommands("wh")
	
	if len(matches) < 1 {
		t.Errorf("expected at least 1 match for 'wh', got %d", len(matches))
		return
	}
	
	// Should find some command starting with 'wh' 
	if len(matches[0]) < 1 {
		t.Errorf("expected non-empty completion suffix for 'wh', got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_completeCommands_Multiple(t *testing.T) {
	evaluator := NewGoEvaluator()
	
	c := NewGoshCompleterForTesting(evaluator)

	// Test multiple matches - should find cd, cat, and possibly PATH executables
	matches := c.completeCommands("c")
	
	expectedMinMatches := 2 // At least cd, cat
	if len(matches) < expectedMinMatches {
		t.Errorf("expected at least %d matches for 'c', got %d", expectedMinMatches, len(matches))
		return
	}
	
	// Check for expected builtin completions
	expectedBuiltinSuffixes := []string{"d", "at"} // cd, cat
	resultStrings := make([]string, len(matches))
	for i, match := range matches {
		resultStrings[i] = string(match)
	}
	
	for _, expectedSuffix := range expectedBuiltinSuffixes {
		found := false
		for _, actual := range resultStrings {
			if actual == expectedSuffix {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected builtin suffix '%s' not found in results: %v", expectedSuffix, resultStrings)
		}
	}
}
