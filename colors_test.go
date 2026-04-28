//go:build darwin || linux

package main

import (
	"strings"
	"testing"
)

func TestGetColorManager(t *testing.T) {
	cm := GetColorManager()

	if cm == nil {
		t.Fatal("GetColorManager returned nil")
	}
}

func TestGetColorManagerSafe(t *testing.T) {
	cm := GetColorManagerSafe()

	if cm == nil {
		t.Fatal("GetColorManagerSafe returned nil")
	}
}

func TestSetYaegiEvalState(t *testing.T) {
	// Test that it doesn't panic
	SetYaegiEvalState(true)
	SetYaegiEvalState(false)
}

func TestStyleMessage(t *testing.T) {
	cm := GetColorManager()

	result := cm.StyleMessage("test message", "welcome")

	if result == "" {
		t.Error("StyleMessage should return non-empty string")
	}

	// Result should contain the message
	if !strings.Contains(result, "test message") {
		t.Errorf("Expected 'test message' in result, got %q", result)
	}
}

func TestSetColorTheme(t *testing.T) {
	// Test setting a theme
	SetColorTheme("dark")

	// Test setting invalid theme (should not panic)
	SetColorTheme("nonexistent_theme")
}

func TestListThemes(t *testing.T) {
	themes := ListThemes()

	if len(themes) == 0 {
		t.Error("ListThemes should return at least one theme")
	}
}

func TestGetCurrentThemeName(t *testing.T) {
	name := GetCurrentThemeName()

	if name == "" {
		t.Error("GetCurrentThemeName should return non-empty string")
	}
}

func TestPrintCurrentTheme(t *testing.T) {
	// Should not panic
	PrintCurrentTheme()
}

func TestExportTheme(t *testing.T) {
	theme := ExportTheme()

	if theme == "" {
		t.Error("ExportTheme should return non-empty string")
	}

	// Should contain theme information (not necessarily JSON)
	if !strings.Contains(theme, "dark") && !strings.Contains(theme, "light") {
		t.Errorf("Expected theme name in result, got %q", theme[:min(50, len(theme))])
	}
}

func TestForceRefresh(t *testing.T) {
	cm := GetColorManager()

	// Should not panic
	cm.ForceRefresh()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
