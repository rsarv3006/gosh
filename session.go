//go:build darwin || linux

package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/traefik/yaegi/interp"
)

type SessionState struct {
	CapturedVars map[string][]string
	GoInterp     *interp.Interpreter
	WorkingDir   string
	Mode         BlockMode
	History      []HistoryBlock
	HistoryFile  string
}

func NewSessionState() *SessionState {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}

	historyFile := filepath.Join(home, ".gosh_history")

	s := &SessionState{
		CapturedVars: make(map[string][]string),
		WorkingDir:   "",
		Mode:         ModeShell,
		History:      []HistoryBlock{},
		HistoryFile:  historyFile,
	}

	wd, err := os.Getwd()
	if err == nil {
		s.WorkingDir = wd
	}

	// Don't load history on startup - avoid displaying old commands
	// s.loadHistory()

	return s
}

func (s *SessionState) loadHistory() {
	data, err := os.ReadFile(s.HistoryFile)
	if err != nil {
		return
	}

	var blocks []HistoryBlock
	if err := json.Unmarshal(data, &blocks); err == nil {
		s.History = blocks
	}
}

func (s *SessionState) saveHistory() error {
	data, err := json.Marshal(s.History)
	if err != nil {
		return err
	}
	return os.WriteFile(s.HistoryFile, data, 0644)
}

func (s *SessionState) AddHistory(block HistoryBlock) {
	s.History = append(s.History, block)
	if len(s.History) > 1000 {
		s.History = s.History[len(s.History)-1000:]
	}
	s.saveHistory()
}

func (s *SessionState) GetPrompt() string {
	if s.Mode == ModeGo {
		return "go> "
	}
	return "$ "
}

func (s *SessionState) GetContinuationPrompt() string {
	return "... "
}
