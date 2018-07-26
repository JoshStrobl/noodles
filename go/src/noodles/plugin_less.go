package main

import (
	"errors"
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"strings"
)

// This is the LESS plugin

type LessPlugin struct {
	CompilerFlags []string
}

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

// Lint will check the specified project's settings related to our plugin
func (p *LessPlugin) Lint(n *NoodlesProject) NoodlesLintResult {
	results := NoodlesLintResult{
		Deprecations:    []string{},
		Errors:          []string{},
		Recommendations: []string{},
	}

	return results
}

// PreRun will check if the necessary lessc executable is installed
func (l *LessPlugin) PreRun(n *NoodlesProject) error {
	var preRunErr error

	if !coreutils.ExecutableExists("lessc") { // If the lessc executable does not exist
		preRunErr = errors.New("lessc is not installed on your system. Please run noodles setup.")
	}

	return preRunErr
}

// PostRun is just a stub function. Doesn't actually do anything at the moment
func (l *LessPlugin) PostRun(n *NoodlesProject) error {
	return nil
}

// Run will compile our LESS into CSS
func (l *LessPlugin) Run(n *NoodlesProject) error {
	var runErr error

	if n.Destination == "" { // If no Destination is set
		n.Destination = filepath.Join("build", n.SimpleName+".css")
	}

	if n.Source == "" { // If no Source is set
		n.Source = filepath.Join("src/less/", n.SimpleName+".less")
	}

	lessFlags := LessCompilerFlags
	lessFlags = append(lessFlags, n.Source, n.Destination) // Add our source and destination to flags

	commandOutput := coreutils.ExecCommand("lessc", lessFlags, false) // Call execCommand and get its commandOutput

	if strings.Contains(commandOutput, "SyntaxError") { // If lessc reported syntax errors
		runErr = errors.New(commandOutput)
	}

	return runErr
}
