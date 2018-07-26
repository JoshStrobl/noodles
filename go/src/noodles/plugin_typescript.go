package main

import (
	"errors"
	"fmt"
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"regexp"
	"strings"
)

// This is the Typescript plugin

type TypeScriptPlugin struct {
}

// SimpleTypescriptCompilerOptions are simple compiler options, namely declaration creation and removal of comments.
var SimpleTypescriptCompilerOptions []string

// AdvancedTypescriptCompilerOptions are advanced compiler options, includes simple options.
var AdvancedTypescriptCompilerOptions []string

// StrictTypescriptCompilerOptions are our most strict compiler options, includes advanced options.
var StrictTypescriptCompilerOptions []string

// ValidTypeScriptModes is a list of valid TypeScript flag modes
var ValidTypeScriptModes []string

// ValidTypeScriptTargets is a list of valid TypeScript targets
var ValidTypeScriptTargets []string

// Do some Typescript compiler option initing
func init() {
	SimpleTypescriptCompilerOptions = []string{
		"--declaration",    // Create a declaration file
		"--removeComments", // Remove comments
	}

	AdvancedTypescriptCompilerOptions = []string{
		"--noFallthroughCasesInSwitch", // Disallow fallthrough cases in switches
		"--noImplicitReturns",          // Disallow implicit returns
	}
	AdvancedTypescriptCompilerOptions = append(AdvancedTypescriptCompilerOptions, SimpleTypescriptCompilerOptions...)

	StrictTypescriptCompilerOptions = []string{
		"--forceConsistentCasingInFileNames", // Enforce consistency in file names
	}

	StrictTypescriptCompilerOptions = append(StrictTypescriptCompilerOptions, AdvancedTypescriptCompilerOptions...)

	ValidTypeScriptModes = []string{"simple", "advanced", "strict"}
	ValidTypeScriptTargets = []string{"ES5", "ES6", "ES7"}
}

// Lint will check the specified project's settings related to our plugin
func (p *TypeScriptPlugin) Lint(n *NoodlesProject) NoodlesLintResult {
	results := NoodlesLintResult{
		Deprecations:    []string{},
		Errors:          []string{},
		Recommendations: []string{},
	}

	if !n.Compress { // Compression not enabled
		results.Recommendations = append(results.Recommendations, "Compression is not enabled, meaning we will only generate a non-minified JS file. Recommended enabling Compress.")
	}

	if n.Mode == "" {
		results.Recommendations = append(results.Recommendations, "No mode is set, meaning we'll default to Advanced flag set. Recommend setting a Mode.")
	} else if !ListContains(ValidTypeScriptModes, n.Mode) {
		results.Errors = append(results.Errors, "No valid Mode set. Must be simple, advanced, or strict.")
	}

	if n.Target == "" {
		results.Recommendations = append(results.Recommendations, "No Target set, meaning we default to ES5. Recommend setting Target to ES5, ES6, or ES7.")
	} else if !ListContains(ValidTypeScriptTargets, n.Target) {
		results.Errors = append(results.Errors, "No valid target set. Must be ES5, ES6, or ES7.")
	}

	return results
}

// PreRun will check if the necessary executables for TypeScript and compression are installed
func (t *TypeScriptPlugin) PreRun(n *NoodlesProject) error {
	var preRunErr error
	executables := []string{"tsc", "uglifyjs2"}

	for _, executable := range executables { // For each executable
		if !coreutils.ExecutableExists(executable) { // If this executable does not exist
			preRunErr = errors.New(executable + " is not installed on your system. Please run noodles setup.")
			break
		}
	}

	return preRunErr
}

// PostRun will perform compression if the project has enabled it
func (t *TypeScriptPlugin) PostRun(n *NoodlesProject) error {
	var postRun error

	if n.Compress { // If we should minify the content
		fmt.Println("Minifying compiled JavaScript.")

		minifiedJSDestination := strings.Replace(n.Destination, ".js", ".min.js", -1) // Replace .js with .min.js
		uglifyArgs := []string{                                                       // Define uglifyArgs
			n.Destination,    // Input
			"--compress",     // Yes, I like to compress things
			"--mangle",       // Mangle variable names
			"warnings=false", // Don't provide warnings
		}

		closureOutput := coreutils.ExecCommand("uglifyjs2", uglifyArgs, true) // Run Google Closure Compiler and store the output in closureOutput
		nodeDeprecationRemover, _ := regexp.Compile(`\(node\:.+\n`)           // Remove any lines starting with (node:
		closureOutput = nodeDeprecationRemover.ReplaceAllString(closureOutput, "")
		closureOutput = strings.TrimSpace(closureOutput) // Fix trailing newlines

		postRun = coreutils.WriteOrUpdateFile(minifiedJSDestination, []byte(closureOutput), coreutils.NonGlobalFileMode) // Write or update the minified JS file content to build/lowercaseProjectName.min.js
	}

	return postRun
}

// Run will run our TypeScript compilation
func (t *TypeScriptPlugin) Run(n *NoodlesProject) error {
	var runErr error

	if n.Destination == "" { // If no custom Destination is set
		n.Destination = filepath.Join("build", n.SimpleName+".js")
	}

	n.Mode = strings.ToLower(n.Mode) // Lowercase n.Mode

	if n.Mode == "" || ((n.Mode != "simple") && (n.Mode != "advanced") && (n.Mode != "strict")) { // If no Mode is set, or is not set to a valid one
		n.Mode = "advanced" // Pick a reasonable middleground
	}

	if n.Source == "" { // If no source is defined
		n.Source = filepath.Join("src", "typescript", n.SimpleName+".ts")
	}

	if !ListContains(ValidTypeScriptTargets, n.Target) { // If this is not a valid target
		n.Target = "ES5" // Set to ES5
	}

	var modeTypeArgs []string // The mode args we'll be using during compilation

	switch n.Mode {
	case "simple":
		modeTypeArgs = SimpleTypescriptCompilerOptions
	case "advanced":
		modeTypeArgs = AdvancedTypescriptCompilerOptions
	case "strict":
		modeTypeArgs = StrictTypescriptCompilerOptions
	}

	tscFlags := append(modeTypeArgs, []string{ // Set tscFlags to the flags we'll pass to the Typescript copmiler
		"--target", n.Target, // Add our target
		"--outFile", n.Destination, // Add our destination
		n.Source, // Add source
	}...)

	commandOutput := coreutils.ExecCommand("tsc", tscFlags, false) // Call execCommand and get its commandOutput

	if strings.Contains(commandOutput, "error TS") { // If tsc reported errors
		runErr = errors.New(commandOutput)
	}

	return runErr
}
