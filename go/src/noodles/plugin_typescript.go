package main

import (
	"fmt"
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"strings"
)

// This is the Typescript plugin

// SimpleTypescriptCompilerOptions are simple compiler options, namely declaration creation and removal of comments.
var SimpleTypescriptCompilerOptions []string

// AdvancedTypescriptCompilerOptions are advanced compiler options, includes simple options.
var AdvancedTypescriptCompilerOptions []string

// StrictTypescriptCompilerOptions are our most strict compiler options, includes advanced options.
var StrictTypescriptCompilerOptions []string

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
}

// Typescript is our plugin functionality for compilation of TypeScript into Javascript.
func (n *NoodlesProject) Typescript(project string) {
	if !coreutils.ExecutableExists("tsc") { // If the tsc executable exists
		fmt.Println("tsc is not installed on your system. Please run noodles setup.")
		return
	}

	if n.Destination == "" { // If no Destination is set
		n.Destination = filepath.Join("build", project+".js")
	}

	if n.Mode == "" || ((n.Mode != "simple") && (n.Mode != "advanced") && (n.Mode != "strict")) { // If no Mode is set, or is not set to a valid one
		n.Mode = "Advanced" // Pick a reasonable middleground
	}

	defaultTargetArgs := []string{"--target", "ES5"}           // Set the default target to ES5
	defaultOutFileArgs := []string{"--outFile", n.Destination} // Set the default outFile to build/{n}.js

	var modeTypeArgs []string // The mode we'll be using during compilation

	switch n.Mode {
	case "simple":
		modeTypeArgs = SimpleTypescriptCompilerOptions
	case "advanced":
		modeTypeArgs = AdvancedTypescriptCompilerOptions
	case "strict":
		modeTypeArgs = StrictTypescriptCompilerOptions
	}

	if len(n.Flags) == 0 { // If no flags are set
		n.Flags = append(defaultTargetArgs, defaultOutFileArgs...) // Set n.Flags to default target and outFile args
	} else { // If flags are set
		joinedFlags := strings.Join(n.Flags, " ")

		if !strings.Contains(joinedFlags, "--target") { // If no target is set
			n.Flags = append(n.Flags, defaultTargetArgs...)
		}

		if !strings.Contains(joinedFlags, "--outFile") { // If no outFile is set
			n.Flags = append(n.Flags, defaultOutFileArgs...)
		}
	}

	n.Flags = append(modeTypeArgs, n.Flags...) // Prepend modeTypeArgs
	n.Flags = append(n.Flags, n.Source)        // Append our source

	commandOutput := coreutils.ExecCommand("tsc", n.Flags, false) // Call execCommand and get its commandOutput

	if !strings.Contains(commandOutput, "error TS") { // If tsc did not report any errors
		if n.Compress { // If we should minify the content
			n.MinifyJavaScript() // Call the minification func, set any error to compileError
		}
	} else { // If tsc did report errors
		fmt.Println(commandOutput)
	}
}

// MinifyJavaScript minifies the JavaScript using Google Closure Compiler and then proceed to attempt to provide a zopfli compressed version.
func (n *NoodlesProject) MinifyJavaScript() {
	if coreutils.ExecutableExists("ccjs") { // If the ccjs executable exists
		fmt.Println("Minifying compiled JavaScript.")

		closureArgs := []string{ // Define closureArgs as
			n.Destination,                              // Yes, I like to compress things.
			"--compilation_level=SIMPLE_OPTIMIZATIONS", // Simple optimizations
			"--warning_level=QUIET",                    // Shush you...
		}

		minifiedJSDestination := strings.Replace(n.Destination, ".js", ".min.js", -1) // Replace .js with .min.js

		closureOutput := coreutils.ExecCommand("ccjs", closureArgs, true)                                      // Run Google Closure Compiler and store the output in closureOutput
		coreutils.WriteOrUpdateFile(minifiedJSDestination, []byte(closureOutput), coreutils.NonGlobalFileMode) // Write or update the minified JS file content to build/lowercaseProjectName.min.js
	} else { // If ccjs does not exist
		fmt.Println("ccjs is not installed. Please run noodles setup.")
	}
}
