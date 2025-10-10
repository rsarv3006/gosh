//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color configuration structures
type PromptColors struct {
	Directory string `json:"directory"`
	GitBranch string `json:"git_branch"`
	Separator string `json:"separator"`
	Symbol    string `json:"symbol"`
}

type OutputColors struct {
	Success string `json:"success"`
	Error   string `json:"error"`
	Info    string `json:"info"`
	Result  string `json:"result"`
}

type MessageColors struct {
	Welcome string `json:"welcome"`
	Config  string `json:"config"`
	Help    string `json:"help"`
}

type ColorTheme struct {
	Name     string       `json:"name"`
	Prompt   PromptColors `json:"prompt"`
	Output   OutputColors `json:"output"`
	Messages MessageColors `json:"messages"`
}

// Built-in theme presets
var builtinThemes = map[string]ColorTheme{
	"dark": {
		Name: "dark",
		Prompt: PromptColors{
			Directory: "#00bcd4", // Cyan
			GitBranch: "#4fc3f7", // Light blue
			Separator: "#607d8b", // Blue gray
			Symbol:    "#ffc107", // Amber
		},
		Output: OutputColors{
			Success: "#4caf50", // Green
			Error:   "#f44336", // Red
			Info:    "#2196f3", // Blue
			Result:  "#ffffff", // White
		},
		Messages: MessageColors{
			Welcome: "#ff9800", // Orange
			Config:  "#9c27b0", // Purple
			Help:    "#607d8b", // Blue gray
		},
	},
	"light": {
		Name: "light",
		Prompt: PromptColors{
			Directory: "#1976d2", // Blue
			GitBranch: "#0288d1", // Darker blue
			Separator: "#757575", // Gray
			Symbol:    "#f57c00", // Dark orange
		},
		Output: OutputColors{
			Success: "#388e3c", // Dark green
			Error:   "#d32f2f", // Dark red
			Info:    "#1976d2", // Blue
			Result:  "#212121", // Dark gray
		},
		Messages: MessageColors{
			Welcome: "#f57c00", // Dark orange
			Config:  "#7b1fa2", // Dark purple
			Help:    "#616161", // Medium gray
		},
	},
	"mono": {
		Name: "mono",
		Prompt: PromptColors{
			Directory: "",  // No color
			GitBranch: "", // No color
			Separator: "", // No color
			Symbol:    "", // No color
		},
		Output: OutputColors{
			Success: "", // No color
			Error:   "", // No color
			Info:    "", // No color
			Result:  "", // No color
		},
		Messages: MessageColors{
			Welcome: "", // No color
			Config:  "", // No color
			Help:    "", // No color
		},
	},
	"solarized": {
		Name: "solarized",
		Prompt: PromptColors{
			Directory: "#268bd2", // Solarized blue
			GitBranch: "#2aa198", // Solarized cyan
			Separator: "#586e75", // Solarized base01
			Symbol:    "#b58900", // Solarized yellow
		},
		Output: OutputColors{
			Success: "#859900", // Solarized green
			Error:   "#dc322f", // Solarized red
			Info:    "#268bd2", // Solarized blue
			Result:  "#839496", // Solarized base0
		},
		Messages: MessageColors{
			Welcome: "#cb4b16", // Solarized orange
			Config:  "#d33682", // Solarized magenta
			Help:    "#586e75", // Solarized base01
		},
	},
}

// ColorManager handles all styling operations
type ColorManager struct {
	theme       ColorTheme
	style       lipgloss.Style
	noColor     bool
	currentName string
}

// Global color manager instance
var colorManager *ColorManager

// Global flag to track when we're in yaegi evaluation
var inYaegiEval = false

// NewColorManager creates a new color manager with default theme
func NewColorManager() *ColorManager {
	return &ColorManager{
		theme:       builtinThemes["dark"], // Default dark theme
		style:       lipgloss.NewStyle(),
		noColor:     shouldUseNoColor(),
		currentName: "dark",
	}
}

// shouldUseNoColor checks if we should disable colors
func shouldUseNoColor() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	
	// Check if stdout is not a TTY
	if !isTerminal() {
		return true
	}
	
	return false
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	// Simple check - in a real implementation you might use more sophisticated detection
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// SetColorTheme sets the current color theme
func SetColorTheme(theme interface{}) {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	
	switch t := theme.(type) {
	case string:
		// Preset theme by name
		if presetTheme, exists := builtinThemes[t]; exists {
			colorManager.theme = presetTheme
			colorManager.currentName = t
		}
		
		// Try dynamic theme creation with hex colors for presets that aren't built-in
		switch t {
		case "light":
			colorManager.theme = createLightTheme()
			colorManager.currentName = t
		case "mono":
			colorManager.theme = createMonoTheme()
			colorManager.currentName = t
		}
		
	case ColorTheme:
		// Custom theme object
		colorManager.theme = t
		colorManager.currentName = t.Name
	}
}

// Helper functions for simple color customization
func SetPromptColor(component, color string) {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	
	switch strings.ToLower(component) {
	case "directory", "dir":
		colorManager.theme.Prompt.Directory = color
	case "git-branch", "git":
		colorManager.theme.Prompt.GitBranch = color
	case "separator", "sep":
		colorManager.theme.Prompt.Separator = color
	case "symbol":
		colorManager.theme.Prompt.Symbol = color
	}
}

func SetOutputColor(outputType, color string) {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	
	switch strings.ToLower(outputType) {
	case "success":
		colorManager.theme.Output.Success = color
	case "error":
		colorManager.theme.Output.Error = color
	case "info":
		colorManager.theme.Output.Info = color
	case "result":
		colorManager.theme.Output.Result = color
	}
}

// StylePrompt styles the prompt components
func (cm *ColorManager) StylePrompt(text, component string) string {
	if cm.noColor || text == "" {
		return text
	}
	
	var color string
	switch component {
	case "directory":
		color = cm.theme.Prompt.Directory
	case "git-branch":
		color = cm.theme.Prompt.GitBranch
	case "separator":
		color = cm.theme.Prompt.Separator
	case "symbol":
		color = cm.theme.Prompt.Symbol
	default:
		return text
	}
	
	if color == "" {
		return text
	}
	
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// StyleOutput styles output based on type
func (cm *ColorManager) StyleOutput(text, outputType string) string {
	if cm.noColor || text == "" {
		return text
	}
	
	var color string
	switch outputType {
	case "success":
		color = cm.theme.Output.Success
	case "error":
		color = cm.theme.Output.Error
	case "info":
		color = cm.theme.Output.Info
	case "result":
		color = cm.theme.Output.Result
	default:
		return text
	}
	
	if color == "" {
		return text
	}
	
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}



// GetColorManager returns the global color manager
// Safe to call anytime (including during yaegi evaluation)
func GetColorManager() *ColorManager {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	return colorManager
}

// SetYaegiEvalState is called by the evaluator to indicate when we're evaluating
func SetYaegiEvalState(inEval bool) {
	inYaegiEval = inEval
}

// StyleMessage safely styles message text, avoiding calls during yaegi eval
func (cm *ColorManager) StyleMessage(text, messageType string) string {
	if cm.noColor || inYaegiEval || text == "" {
		return text
	}
	
	var color string
	switch messageType {
	case "welcome":
		color = cm.theme.Messages.Welcome
	case "config":
		color = cm.theme.Messages.Config
	case "help":
		color = cm.theme.Messages.Help
	default:
		return text
	}
	
	if color == "" {
		return text
	}
	
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}

// GetColorManagerSafe returns the global color manager
// Only use when NOT during yaegi evaluation
func GetColorManagerSafe() *ColorManager {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	return colorManager
}

// ListThemes returns available theme names
func ListThemes() []string {
	themes := make([]string, 0, len(builtinThemes))
	for name := range builtinThemes {
		themes = append(themes, name)
	}
	return themes
}

// GetCurrentThemeName returns the current theme name
func GetCurrentThemeName() string {
	if colorManager == nil {
		return "dark"
	}
	return colorManager.currentName
}

// PrintCurrentTheme prints the current theme color values
func PrintCurrentTheme() {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	
	theme := colorManager.theme
	fmt.Printf("Current theme: %s\n", theme.Name)
	fmt.Printf("Prompt Colors:\n")
	fmt.Printf("  Directory: %s\n", theme.Prompt.Directory)
	fmt.Printf("  GitBranch: %s\n", theme.Prompt.GitBranch)
	fmt.Printf("  Separator: %s\n", theme.Prompt.Separator)
	fmt.Printf("  Symbol: %s\n", theme.Prompt.Symbol)
	fmt.Printf("Output Colors:\n")
	fmt.Printf("  Success: %s\n", theme.Output.Success)
	fmt.Printf("  Error: %s\n", theme.Output.Error)
	fmt.Printf("  Info: %s\n", theme.Output.Info)
	fmt.Printf("  Result: %s\n", theme.Output.Result)
	fmt.Printf("Message Colors:\n")
	fmt.Printf("  Welcome: %s\n", theme.Messages.Welcome)
	fmt.Printf("  Config: %s\n", theme.Messages.Config)
	fmt.Printf("  Help: %s\n", theme.Messages.Help)
}

// ExportTheme returns a string representation of the current theme for copy-pasting
func ExportTheme() string {
	if colorManager == nil {
		colorManager = NewColorManager()
	}
	
	theme := colorManager.theme
	return fmt.Sprintf(`ColorTheme{
	Name: "%s",
	Prompt: PromptColors{
		Directory: "%s",
		GitBranch: "%s",
		Separator: "%s",
		Symbol:    "%s",
	},
	Output: OutputColors{
		Success: "%s",
		Error:   "%s",
		Info:    "%s",
		Result:  "%s",
	},
	Messages: MessageColors{
		Welcome: "%s",
		Config:  "%s",
		Help:    "%s",
	},
}`, theme.Name,
		theme.Prompt.Directory, theme.Prompt.GitBranch, theme.Prompt.Separator, theme.Prompt.Symbol,
		theme.Output.Success, theme.Output.Error, theme.Output.Info, theme.Output.Result,
		theme.Messages.Welcome, theme.Messages.Config, theme.Messages.Help)
}

// Helper function to create light theme dynamically
func createLightTheme() ColorTheme {
	return ColorTheme{
		Name: "light",
		Prompt: PromptColors{
			Directory: "#1976d2",
			GitBranch: "#0288d1",
			Separator: "#757575",
			Symbol:    "#f57c00",
		},
		Output: OutputColors{
			Success: "#388e3c",
			Error:   "#d32f2f",
			Info:    "#1976d2",
			Result:  "#212121",
		},
		Messages: MessageColors{
			Welcome: "#f57c00",
			Config:  "#7b1fa2",
			Help:    "#616161",
		},
	}
}

// Helper function to create mono theme dynamically
func createMonoTheme() ColorTheme {
	return ColorTheme{
		Name: "mono",
		Prompt: PromptColors{
			Directory: "",
			GitBranch: "",
			Separator: "",
			Symbol:    "",
		},
		Output: OutputColors{
			Success: "",
			Error:   "",
			Info:    "",
			Result:  "",
		},
		Messages: MessageColors{
			Welcome: "",
			Config:  "",
			Help:    "",
		},
	}
}
