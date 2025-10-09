//go:build darwin || linux

package main

type InputType int

const (
	InputTypeGo InputType = iota
	InputTypeCommand
	InputTypeBuiltin
)

type ExecutionResult struct {
	Output   string
	ExitCode int
	Error    error
}
