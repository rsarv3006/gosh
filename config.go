// Local config - loads second, overrides home config
package main

import (
	"fmt"
	"os"
)

func init() {
	fmt.Println("LOCAL config: Setting up project-specific environment...")
	// Override home config's EDITOR setting
	os.Setenv("EDITOR", "code")
	// Add local-only env vars
	os.Setenv("GOSH_SOURCE", "local-config")
	os.Setenv("GOSH_USER", "config_loaded")
	fmt.Println("LOCAL config: Defining project-specific functions...")
}

// Override home config's info function
func info() {
	fmt.Printf("LOCAL config: Project-specific info - EDITOR=%s, SOURCE=%s\n", os.Getenv("EDITOR"), os.Getenv("GOSH_SOURCE"))
}

// Local-only function
func projectCmd() {
	fmt.Println("LOCAL config: Project-specific command")
}
