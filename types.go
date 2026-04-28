//go:build darwin || linux

package main

type BlockMode int

const (
	ModeShell BlockMode = iota
	ModeGo
)

type HistoryBlock struct {
	Mode    BlockMode
	Input   string
	Output  string
	Capture string // Variable name if -> was used, empty otherwise
}

type InputType int

const (
	InputTypeGo InputType = iota
	InputTypeCommand
	InputTypeBuiltin
	InputTypeModeSwitch
)

type ExecutionResult struct {
	Output   string
	ExitCode int
	Error    error
}
