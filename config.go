package main

import (
	"fmt"
	"os"
)

// Shell startup configuration
func init() {
	// Environment setup
	fmt.Println("Config: Setting up environment...")
	os.Setenv("GOSH_USER", "config_loaded")
	
	// Custom shell functions available in the REPL
	fmt.Println("Config: Defining custom functions...")
}

// Sample function that can be called from the shell
func hello(name string) {
	fmt.Printf("Hello, %s! Welcome to gosh with config support!\n", name)
}

// Simple function that doesn't use shell APIs yet
func info() {
	fmt.Printf("Config loaded successfully!\n")
	fmt.Printf("User: %s\n", os.Getenv("GOSH_USER"))
}
