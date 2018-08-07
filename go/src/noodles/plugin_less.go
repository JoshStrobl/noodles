package main

import (
	"errors"
	"github.com/stroblindustries/coreutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// LessPlugin is our LESS plugin
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
	results := make(NoodlesLintResult)
	return results
}

// PreRun will check if the necessary lessc executable is installed
func (p *LessPlugin) PreRun(n *NoodlesProject) error {
	var preRunErr error

	if !coreutils.ExecutableExists("lessc") { // If the lessc executable does not exist
		preRunErr = errors.New("lessc is not installed on your system. Please run noodles setup")
	}

	return preRunErr
}

// PostRun is just a stub function. Doesn't actually do anything at the moment
func (p *LessPlugin) PostRun(n *NoodlesProject) error {
	var postRunErr error

	if n.AppendHash { // If we should append the hash
		var fileContent []byte
		fileContent, postRunErr = ioutil.ReadFile(n.Destination)

		if postRunErr == nil { // No error during read
			hash := CreateHash(fileContent)
			fileNameWithoutExtension := strings.Replace(filepath.Base(n.Destination), filepath.Ext(n.Destination), "", -1) // Get the base name and remove the extension
			newFileName := filepath.Join(filepath.Dir(n.Destination), fileNameWithoutExtension+"-"+hash+".css")
			os.Rename(n.Destination, newFileName)
		}
	}

	return postRunErr
}

// Run will compile our LESS into CSS
func (p *LessPlugin) Run(n *NoodlesProject) error {
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
