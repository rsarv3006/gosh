package main

import (
	"fmt"
	"os"
)

// Global debug flag - set to true to enable debug logging
var debug = false

func debugf(format string, args ...interface{}) {
	if !debug {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func debugln(args ...interface{}) {
	if !debug {
		return
	}
	fmt.Fprintln(os.Stderr, args...)
}
