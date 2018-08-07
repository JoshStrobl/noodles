package main

import (
	"errors"
	"fmt"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strings"
)

// GoPlugin is our Go plugin
type GoPlugin struct {
}

var originalGoPath string

// Lint will check the specified project's settings related to our plugin
func (p *GoPlugin) Lint(n *NoodlesProject) NoodlesLintResult {
	results := NoodlesLintResult{
		Deprecations:    []string{},
		Errors:          []string{},
		Recommendations: []string{},
	}

	if !strings.HasSuffix(n.Source, "*.go") { // Globbing isn't enabled
		results.Recommendations = append(results.Recommendations, "Not using globbing for getting all Go files in this project. Recommend changing Sources to *.go.")
	}

	return results
}

// PreRun will check if the necessary Go executable is installed
func (p *GoPlugin) PreRun(n *NoodlesProject) error {
	var preRunErr error

	if !coreutils.ExecutableExists("go") { // If the go executable does not exist
		preRunErr = errors.New("Go is not installed on your system. Please run noodles setup")
	} else { // If the go executable exists
		ToggleGoEnv(true) // Enable the Go environment
	}

	return preRunErr
}

// PostRun will reset our Go environment post-compilation
func (p *GoPlugin) PostRun(n *NoodlesProject) error {
	return ToggleGoEnv(false)
}

// Run will compile the provided project
func (p *GoPlugin) Run(n *NoodlesProject) error {
	var runErr error
	os.Chdir(filepath.Join(workdir, "go")) // Change to our go directory

	if n.Destination == "" { // If no destination is set
		n.Destination = filepath.Join("build", n.SimpleName) // Set destination to build/name (as binary)
	}

	n.Destination = filepath.Join(workdir, n.Destination)

	runErr = os.MkdirAll(filepath.Dir(n.Destination), coreutils.NonGlobalFileMode)

	if runErr == nil { // If there wasn't any error creating the necessary directories
		files := n.GetFiles()
		args := []string{"build", "-o", n.Destination}
		args = append(args, files...)

		goCompilerOutput := coreutils.ExecCommand("go", args, true)

		if strings.Contains(goCompilerOutput, ".go") || strings.Contains(goCompilerOutput, "# command") { // If running the go build shows there are obvious issues
			runErr = errors.New(strings.TrimSpace(goCompilerOutput))
		} else { // If there was no obvious issues
			fmt.Println("Build successful.")
			os.Chdir(filepath.Dir(n.Source))
			coreutils.ExecCommand("gofmt", []string{"-s", "-w", "*"}, true) // Run formatting
		}
	} else { // If we failed to create the necessary directories
		fmt.Printf("Failed to create the necessary directories:\n%s\n", runErr.Error())
		ToggleGoEnv(false)
	}

	return runErr
}

// ToggleGoEnv will toggle our usage of GOPATH and working directory
func ToggleGoEnv(on bool) error {
	var toggleEnvErr error
	if on {
		originalGoPath = os.Getenv("GOPATH") // Store the original GOPATH
		toggleEnvErr = os.Setenv("GOPATH", filepath.Join(workdir, "go"))
	} else {
		if toggleEnvErr = os.Setenv("GOPATH", originalGoPath); toggleEnvErr == nil { // If there was no error setting the environment
			toggleEnvErr = os.Chdir(workdir)
		}
	}

	return toggleEnvErr
}
