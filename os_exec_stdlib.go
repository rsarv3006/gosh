// Manually created os/exec stdlib symbols for yaegi
package main

import (
	"os/exec"
	"reflect"
)

// InjectOSExecSymbols adds os/exec symbols to yaegi's stdlib
func InjectOSExecSymbols() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"os/exec": {
			"Command":  reflect.ValueOf(exec.Command),
			"LookPath": reflect.ValueOf(exec.LookPath),
		},
	}
}
