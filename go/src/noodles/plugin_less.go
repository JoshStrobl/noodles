package main

import (
	"fmt"
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"strings"
)

// This is the LESS plugin

// LessCompilerFlags are the flags we pass to lessc
var LessCompilerFlags []string

func init() {
	LessCompilerFlags = []string{
		"--clean-css",
		"--glob",
		"--no-color",
		"--no-ie-compat",
		"--no-js",
		"--strict-math=on",
	}
}

// LESS is our plugin functionality for compilation of LESS into CSS.
func (n *NoodlesProject) LESS(project string) {
	if !coreutils.ExecutableExists("lessc") { // If the lessc executable does not exist
		fmt.Println("lessc is not installed on your system. Please run noodles setup.")
		return
	}

	if n.Destination == "" { // If no Destination is set
		n.Destination = filepath.Join("build", project+".css")
	}

	if n.Source == "" { // If no Source is set
		n.Source = filepath.Join("src/less/", project+".less")
	}

	lessFlags := LessCompilerFlags                         // Set lessFlags to our LessCompilerFlags
	lessFlags = append(lessFlags, n.Source, n.Destination) // Add our source and destination

	commandOutput := coreutils.ExecCommand("lessc", lessFlags, false) // Call execCommand and get its commandOutput

	if strings.Contains(commandOutput, "SyntaxError") { // If lessc reported syntax errors
		fmt.Println(commandOutput)
	}
}
