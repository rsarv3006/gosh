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

// Simple function that doesn't use shell APIs yet
func info() {
	fmt.Printf("Config loaded successfully!\n")
	fmt.Printf("User: %s\n", os.Getenv("GOSH_USER"))
}
