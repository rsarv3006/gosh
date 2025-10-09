//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// NewGoshCompleter creates a new readline completer
func NewGoshCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(
		// Built-in commands
		readline.PcItem("cd",
			readline.PcItemDynamic(listDynamicDirectories),
		),
		readline.PcItem("pwd"),
		readline.PcItem("exit"),
		readline.PcItem("help",
			readline.PcItem("cd"),
			readline.PcItem("pwd"),
			readline.PcItem("exit"),
			readline.PcItem("help"),
			readline.PcItem("go"),
			readline.PcItem("golang"),
			readline.PcItem("yaegi"),
			readline.PcItem("substitution"),
			readline.PcItem("command"),
		),
		
		// Common shell commands
		readline.PcItem("ls",
			readline.PcItemDynamic(listDynamicFiles),
		),
		readline.PcItem("cat",
			readline.PcItemDynamic(listDynamicFiles),
		),
		readline.PcItem("cd",
			readline.PcItemDynamic(listDynamicDirectories),
		),
		readline.PcItem("grep"),
		readline.PcItem("find"),
		readline.PcItem("ps"),
		readline.PcItem("kill"),
		readline.PcItem("echo"),
		readline.PcItem("date"),
		readline.PcItem("whoami"),
		
		// Go keywords
		readline.PcItem("func"),
		readline.PcItem("var"),
		readline.PcItem("const"),
		readline.PcItem("type"),
		readline.PcItem("import",
			readline.PcItem("fmt"),
			readline.PcItem("os"),
			readline.PcItem("strings"),
			readline.PcItem("strconv"),
			readline.PcItem("path/filepath"),
			readline.PcItem("net/http"),
			readline.PcItem("encoding/json"),
		),
		readline.PcItem("for"),
		readline.PcItem("if"),
		readline.PcItem("switch"),
		readline.PcItem("select"),
		readline.PcItem("return"),
		readline.PcItem("go"),
		readline.PcItem("defer"),
		
		// Dynamic file completion for any input
		readline.PcItemDynamic(listDynamicFiles),
	)
}

// listDynamicFiles returns a function that can list files for completion
func listDynamicFiles(prefix string) []string {
	return listFiles(prefix, false)
}

// listDynamicDirectories returns a function that can list directories for completion
func listDynamicDirectories(prefix string) []string {
	return listFiles(prefix, true)
}

// listFiles lists files or directories matching prefix
func listFiles(prefix string, dirsOnly bool) []string {
	var matches []string
	
	// Extract directory and file pattern
	dir := "."
	pattern := prefix
	
	if strings.LastIndex(prefix, "/") != -1 {
		dir = filepath.Dir(prefix)
		pattern = filepath.Base(prefix)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return matches
	}

	for _, file := range files {
		if dirsOnly && !file.IsDir() {
			continue
		}
		
		name := file.Name()
		if strings.HasPrefix(name, pattern) {
			matches = append(matches, name)
		}
	}

	return matches
}
