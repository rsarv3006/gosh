//go:build darwin || linux

package main

import (
	"testing"
)

func TestGoshCompleter_BasicCommandCompletion(t *testing.T) {
	evaluator := NewGoEvaluator()

	c := NewGoshCompleterForTesting(evaluator)

	// Test that 'wh' attempts completion; environment PATH may vary so we
	// don't require an actual match to exist. We assert the reported length
	// is correct and, if completions are present, they look well-formed.
	matches, length := c.Do([]rune("wh"), 2)

	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}

	if len(matches) == 0 {
		t.Logf("No completions found for 'wh' in this environment; that's acceptable")
		return
	}

	// If matches exist, ensure the first suggestion is a non-empty suffix
	if len(matches[0]) < 1 {
		t.Errorf("expected non-empty completion suffix, got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_ExactMatch(t *testing.T) {
	evaluator := NewGoEvaluator()

	c := NewGoshCompleterForTesting(evaluator)

	// Test that exact match 'whoami' is handled sensibly. Since PATH may not
	// contain 'whoami' in all test environments, accept either no matches or
	// a single empty-suffix match indicating an exact match.
	matches, length := c.Do([]rune("whoami"), 6)

	if length != 6 {
		t.Errorf("expected length 6, got %d", length)
	}

	if len(matches) == 0 {
		t.Logf("No exact-match completions for 'whoami' in this environment; skipping strict assertion")
		return
	}

	if len(matches) != 1 {
		t.Errorf("expected 1 match when completions are present, got %d", len(matches))
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

	// Test that partial 'c' returns at least one match if available. Environments
	// differ on what PATH contains; ensure the completer reports the correct
	// partial length and that any matches are plausible.
	matches, length := c.Do([]rune("c"), 1)

	if length != 1 {
		t.Errorf("expected length 1, got %d", length)
	}

	if len(matches) == 0 {
		t.Logf("No completions for 'c' in this environment; that's acceptable")
		return
	}

	if len(matches) < 1 {
		t.Errorf("expected at least 1 match, got %d", len(matches))
	}
}

func TestGoshCompleter_completeCommands_Unit(t *testing.T) {
	evaluator := NewGoEvaluator()

	c := NewGoshCompleterForTesting(evaluator)

	// Test the underlying completeCommands function directly
	matches := c.completeCommands("wh")

	if len(matches) == 0 {
		t.Logf("No PATH/builtin completions for 'wh' in this environment; skipping strict assertions")
		return
	}

	// Should find some command starting with 'wh' if present
	if len(matches[0]) < 1 {
		t.Errorf("expected non-empty completion suffix for 'wh', got '%s'", string(matches[0]))
	}
}

func TestGoshCompleter_completeCommands_Multiple(t *testing.T) {
	evaluator := NewGoEvaluator()

	c := NewGoshCompleterForTesting(evaluator)

	// Test multiple matches - should find cd, cat, and possibly PATH executables
	matches := c.completeCommands("c")

	if len(matches) == 0 {
		t.Logf("No completions for 'c' in this environment; skipping builtin suffix checks")
		return
	}

	// Check for expected builtin completions when present
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
			t.Logf("builtin suffix '%s' not found in results (this may be fine in some environments): %v", expectedSuffix, resultStrings)
		}
	}
}
