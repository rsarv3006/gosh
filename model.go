//go:build darwin || linux

package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	session    *SessionState
	evaluator  *GoEvaluator
	spawner    *ProcessSpawner
	builtins   *BuiltinHandler
	output     string
	quitting   bool
	width      int
	height     int
	historyIdx int
}

func initialModel(session *SessionState, evaluator *GoEvaluator, spawner *ProcessSpawner, builtins *BuiltinHandler) *model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Focus()
	ta.Prompt = session.GetPrompt()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(1)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(80, 20)

	return &model{
		textarea:   ta,
		viewport:   vp,
		session:    session,
		evaluator:  evaluator,
		spawner:    spawner,
		builtins:   builtins,
		output:     "",
		quitting:   false,
		width:      80,
		height:     24,
		historyIdx: -1,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()
		case tea.KeyUp, tea.KeyDown:
			return m.handleHistory(msg.Type)
		case tea.KeyPgUp:
			m.viewport.LineUp(10)
			return m, nil
		case tea.KeyPgDown:
			m.viewport.LineDown(10)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	m.textarea.Prompt = m.session.GetPrompt()

	// Pass mouse events to viewport for scrolling
	if _, ok := msg.(tea.MouseMsg); ok {
		m.viewport, _ = m.viewport.Update(msg)
	}

	return m, cmd
}

// updateViewportContent rebuilds the viewport content from history
func (m *model) updateViewportContent() {
	var content strings.Builder

	for _, block := range m.session.History {
		if block.Input != "" {
			content.WriteString(m.session.GetPromptForMode(block.Mode) + block.Input + "\n")
		}
		if block.Output != "" {
			content.WriteString(block.Output)
			if !strings.HasSuffix(block.Output, "\n") {
				content.WriteString("\n")
			}
		}
	}

	m.viewport.SetContent(content.String())
}

func (m *model) handleEnter() (tea.Model, tea.Cmd) {
	input := m.textarea.Value()
	if input == "" {
		return m, nil
	}

	m.textarea.Reset()
	m.textarea.Prompt = m.session.GetPrompt()
	m.historyIdx = -1

	// Handle mode switching commands
	if input == ":go" {
		m.session.Mode = ModeGo
		m.textarea.Prompt = m.session.GetPrompt()
		return m, nil
	}
	if input == ":sh" {
		m.session.Mode = ModeShell
		m.textarea.Prompt = m.session.GetPrompt()
		return m, nil
	}

	// Check if input is complete (for multiline Go)
	if m.session.Mode == ModeGo && !isComplete(input) {
		m.textarea.SetValue(input + "\n")
		m.textarea.CursorEnd()
		return m, nil
	}

	// Execute the block
	output := m.executeBlock(input)
	m.output = output

	return m, nil
}

func (m *model) handleHistory(keyType tea.KeyType) (tea.Model, tea.Cmd) {
	if len(m.session.History) == 0 {
		return m, nil
	}

	if keyType == tea.KeyUp {
		if m.historyIdx < len(m.session.History)-1 {
			m.historyIdx++
		}
	} else {
		if m.historyIdx > 0 {
			m.historyIdx--
		} else {
			return m, nil
		}
	}

	idx := len(m.session.History) - 1 - m.historyIdx
	m.textarea.SetValue(m.session.History[idx].Input)
	m.textarea.CursorEnd()

	return m, nil
}

func (m *model) executeBlock(input string) string {
	var result ExecutionResult
	var capturedVar string

	// Check for -> capture syntax
	if m.session.Mode == ModeShell {
		if idx := strings.Index(input, " -> "); idx != -1 {
			capturedVar = strings.TrimSpace(input[idx+3:])
			input = strings.TrimSpace(input[:idx])
		}
	}

	// Route and execute based on mode
	if m.session.Mode == ModeGo {
		result = m.evaluator.EvalWithRecovery(input)
	} else {
		// Shell mode - check for builtins first
		router := NewRouter(m.builtins, &ShellState{WorkingDirectory: m.session.WorkingDir})
		inputType, command, args := router.Route(input)

		switch inputType {
		case InputTypeBuiltin:
			result = m.builtins.Execute(command, args)
		case InputTypeCommand:
			result = m.spawner.ExecuteInteractive(command, args)
		default:
			result = ExecutionResult{Output: fmt.Sprintf("Unknown command: %s\n", command), ExitCode: 1}
		}
	}

	// Handle captured output
	if capturedVar != "" && result.ExitCode == 0 {
		lines := strings.Split(strings.TrimSpace(result.Output), "\n")
		m.session.CapturedVars[capturedVar] = lines
		// Inject into Go interpreter
		m.evaluator.InjectVariable(capturedVar, lines)
		return ""
	}

	// Add to history
	block := HistoryBlock{
		Mode:    m.session.Mode,
		Input:   input,
		Output:  result.Output,
		Capture: capturedVar,
	}
	m.session.AddHistory(block)
	m.updateViewportContent()

	return ""
}

func (m *model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	return m.viewport.View() + "\n" + m.textarea.View()
}

var (
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// isComplete checks if the input is syntactically complete (for multiline Go)
func isComplete(input string) bool {
	input = strings.TrimSpace(input)

	// Check for unclosed braces
	openBraces := strings.Count(input, "{")
	closeBraces := strings.Count(input, "}")
	if openBraces != closeBraces {
		return false
	}

	// Check for unclosed parentheses
	openParens := strings.Count(input, "(")
	closeParens := strings.Count(input, ")")
	if openParens != closeParens {
		return false
	}

	// Check for unclosed brackets
	openBrackets := strings.Count(input, "[")
	closeBrackets := strings.Count(input, "]")
	if openBrackets != closeBrackets {
		return false
	}

	// Check if line ends with incomplete statement
	if strings.HasSuffix(input, ",") ||
		strings.HasSuffix(input, "+") ||
		strings.HasSuffix(input, "-") ||
		strings.HasSuffix(input, "*") ||
		(strings.HasSuffix(input, "/") && !looksLikePathCompletion(input)) ||
		strings.HasSuffix(input, "||") ||
		strings.HasSuffix(input, "&&") {
		return false
	}

	return true
}

// looksLikePathCompletion checks if the trailing "/" is likely from path completion
func looksLikePathCompletion(input string) bool {
	input = strings.TrimSpace(input)

	if !strings.HasSuffix(input, "/") {
		return false
	}

	words := strings.Fields(input)
	if len(words) == 0 {
		return false
	}

	lastWord := words[len(words)-1]

	// If the last word contains path separators or starts with tilde, it's likely a path
	if strings.Contains(lastWord, "/") || strings.HasPrefix(lastWord, "~") {
		return true
	}

	// If there's only one word that ends with "/" and no Go syntax, it's likely a path
	if len(words) == 1 && !strings.ContainsAny(input, "{}();:=") {
		return true
	}

	return false
}
