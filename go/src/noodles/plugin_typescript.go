package main

import (
	"errors"
	"github.com/JoshStrobl/trunk"
	"github.com/stroblindustries/coreutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TypeScriptPlugin is our TypeScript plugin
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
		"--noUnusedLocals",             // Disallow unused locals
		"--noUnusedParameters",         // Disallow unused parameters
	}
	AdvancedTypescriptCompilerOptions = append(AdvancedTypescriptCompilerOptions, SimpleTypescriptCompilerOptions...)

	StrictTypescriptCompilerOptions = []string{
		"--forceConsistentCasingInFileNames", // Enforce consistency in file names
	}

	StrictTypescriptCompilerOptions = append(StrictTypescriptCompilerOptions, AdvancedTypescriptCompilerOptions...)

	ValidTypeScriptModes = []string{"simple", "advanced", "strict"}
	ValidTypeScriptTargets = []string{"ES2017", "ES2018", "ES2019", "ES2020", "ESNext"}
}

// Check will check the specified project's settings related to our plugin
func (p *TypeScriptPlugin) Check(n *NoodlesProject) NoodlesCheckResult {
	results := make(NoodlesCheckResult)

	deprecations := []string{}
	errors := []string{}
	recommendations := []string{}

	if !n.Compress { // Compression not enabled
		recommendations = append(recommendations, "Compression is not enabled, meaning we will only generate a non-minified JS file. Recommended enabling Compress.")
	}

	if n.Mode == "" {
		recommendations = append(recommendations, "No mode is set, meaning we'll default to Advanced flag set. Recommend setting a Mode.")
	} else if !ListContains(ValidTypeScriptModes, n.Mode) {
		errors = append(errors, "No valid Mode set. Must be simple, advanced, or strict.")
	}

	if n.Target == "" {
		recommendations = append(recommendations, "No Target set, meaning we default to the latest formal specification, currently ES2018.")
	} else if !ListContains(ValidTypeScriptTargets, n.Target) {
		errors = append(errors, "No valid target set. Must be "+strings.Join(ValidTypeScriptTargets, ", "))
	}

	results["Deprecations"] = deprecations
	results["Errors"] = errors
	results["Recommendations"] = recommendations

	return results
}

// Lint is currently a stub func, offers no functionality yet.
func (p *TypeScriptPlugin) Lint(n *NoodlesProject, confidence float64) (lintErr error) {
	trunk.LogErr("Linting of TypeScript projects not currently supported.")
	return
}

// PreRun will check if the necessary executables for TypeScript and compression are installed
func (p *TypeScriptPlugin) PreRun(n *NoodlesProject) (preRunErr error) {
	executables := []string{DependenciesMap["compress"].Binary, DependenciesMap["typescript"].Binary}

	for _, executable := range executables { // For each executable
		if !coreutils.ExecutableExists(executable) { // If this executable does not exist
			preRunErr = errors.New(executable + " is not installed on your system. Please run noodles setup.")
			break
		}
	}

	return
}

// PostRun will perform compression if the project has enabled it
func (p *TypeScriptPlugin) PostRun(n *NoodlesProject) (postRunErr error) {
	destDir := filepath.Dir(n.Destination)
	fileName := filepath.Base(n.Destination)
	fileNameWithoutExtension := strings.Replace(fileName, filepath.Ext(n.Destination), "", -1) // Get the base name and remove the extension

	if n.AppendHash { // Appended hash
		jsExtension := ".js"

		if n.Compress { // If we're not compressing
			jsExtension = ".min.js"
		}

		RemoveHashedFiles(destDir, jsExtension, fileNameWithoutExtension) // Remove existing hashed files
	}

	if n.Compress { // If we should minify the content
		trunk.LogInfo("Minifying compiled JavaScript.")

		uglifyArgs := []string{ // Define uglifyArgs
			n.Destination, // Input
			"--compress",  // Yes, I like to compress things
			"--mangle",    // Mangle variable names
		}

		closureOutput := coreutils.ExecCommand("terser", uglifyArgs, true) // Run our JavaScript compressor / minifier and store the output in closureOutput
		nodeDeprecationRemover, _ := regexp.Compile(`\(node\:.+\n`)        // Remove any lines starting with (node:
		closureOutput = nodeDeprecationRemover.ReplaceAllString(closureOutput, "")
		closureOutput = strings.TrimSpace(closureOutput) // Fix trailing newlines

		var minifiedJSDestination string

		if n.AppendHash { // If we should append the hash, just immediately set our minifiedJSDestination so we can skip our move step
			hash := CreateHash([]byte(closureOutput))
			minifiedJSDestination = filepath.Join(destDir, fileNameWithoutExtension+"-"+hash+".min.js")
		} else {
			minifiedJSDestination = filepath.Join(destDir, fileNameWithoutExtension+".min.js")
		}

		postRunErr = coreutils.WriteOrUpdateFile(minifiedJSDestination, []byte(closureOutput), coreutils.NonGlobalFileMode) // Write or update the minified JS file content to build/lowercaseProjectName.min.js
	} else { // If we're not minifying the content
		if n.AppendHash { // If we're appending the hash to the .js file
			var fileContent []byte
			fileContent, postRunErr = ioutil.ReadFile(n.Destination)

			if postRunErr == nil { // No error during read
				hash := CreateHash(fileContent)
				newFileName := filepath.Join(destDir, fileNameWithoutExtension+"-"+hash+".js")
				os.Rename(n.Destination, newFileName) // Rename the file
			}
		}
	}

	return
}

// RequiresPreRun is a stub function.
func (p *TypeScriptPlugin) RequiresPreRun(n *NoodlesProject) error {
	return nil
}

// RequiresPostRun is a stub function.
func (p *TypeScriptPlugin) RequiresPostRun(n *NoodlesProject) error {
	return nil
}

// Run will run our TypeScript compilation
func (p *TypeScriptPlugin) Run(n *NoodlesProject) (runErr error) {
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
		n.Target = "ES2019" // Set to 2019
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

	commandOutput := coreutils.ExecCommand("tsc", tscFlags, true) // Call execCommand and get its commandOutput

	if strings.Contains(commandOutput, "error TS") { // If tsc reported errors
		runErr = errors.New(commandOutput)
	}

	return
}
