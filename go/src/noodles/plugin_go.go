package main

import (
	"errors"
	"fmt"
	"github.com/stroblindustries/coreutils"
	xlint "golang.org/x/lint"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GoPlugin is our Go plugin
type GoPlugin struct {
}

var originalGoPath string

// Check will check the specified project's settings related to our plugin
func (p *GoPlugin) Check(n *NoodlesProject) NoodlesCheckResult {
	results := make(NoodlesCheckResult)
	recommendations := []string{}

	if n.Type == "" { // No type designated
		recommendations = append(recommendations, "Not setting any type. Will default to binary. Recommend statically setting this.")
	}

	if !strings.HasSuffix(n.Source, "*.go") { // Globbing isn't enabled
		recommendations = append(recommendations, "Not using globbing for getting all Go files in this project. Recommend changing Sources to *.go.")
	}

	if len(recommendations) != 0 {
		results["Recommendations"] = recommendations
	}

	return results
}

// Lint will lint our Go code. Lint takes a Noodles Project and the minimum acceptable confidence
func (p *GoPlugin) Lint(n *NoodlesProject, confidence float64) error {
	var lintErr error
	sourceDir := filepath.Dir(n.Source)

	goFiles, getErr := coreutils.GetFilesContains(sourceDir, ".go") // Get all files with .go extension

	if getErr == nil { // Get all files with .go extension
		if len(goFiles) != 0 { // If we managed to find files

			for _, fileName := range goFiles { // For each file
				fileContent, readErr := ioutil.ReadFile(fileName)

				if readErr == nil { // If there was no error reading this file
					var linter xlint.Linter
					problems, xlintErr := linter.Lint(fileName, fileContent)

					if xlintErr == nil { // If there was no error during linting
						if len(problems) > 0 {
							for _, problem := range problems { // For each problem
								if problem.Confidence >= confidence { // If the linting confidence is equal to or greater than our requested minimum confidence
									// Example: test.go:24:34: test = errors.New("Hello world.")
									// error strings should not be capitalized or end with punctuation or a newline
									lineErr := strings.TrimSpace(problem.LineText) // Trim any spacing
									fmt.Printf("%s:%d:%d: %s\n%s\n", fileName, problem.Position.Line, problem.Position.Column, lineErr, problem.Text)
								}
							}
						}
					} else {
						lintErr = errors.New("failed to lint " + fileName + ": " + xlintErr.Error())
						break
					}
				} else { // If we failed to read the file
					lintErr = errors.New("failed to read " + fileName + ": " + readErr.Error())
					break
				}
			}
		}
	} else {
		lintErr = errors.New("failed to get files: " + getErr.Error())
	}

	return lintErr
}

// PreRun will check if the necessary Go executable is installed
func (p *GoPlugin) PreRun(n *NoodlesProject) error {
	var preRunErr error

	if !coreutils.ExecutableExists("go") { // If the go executable does not exist
		preRunErr = errors.New("go is not installed on your system. Please run noodles setup")
	} else { // If the go executable exists
		preRunErr = ToggleGoEnv(true) // Enable the Go environment
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

	if n.Destination == "" { // If a destination is set
		if n.Type == "binary" { // If this is a binary
			n.Destination = filepath.Join(workdir, "build", n.SimpleName) // Set destination to build/name (as binary)
		} else if n.Type == "package" { // Package
			n.Destination = workdir
		} else if n.Type == "plugin" { // Plugin
			n.Destination = filepath.Join(workdir, "build", n.SimpleName, ".so") // Set destination to build/name.so
		}
	} else {
		if (n.Type == "plugin") && (filepath.Ext(n.Destination) != ".so") { // Destination does not have .so
			n.Destination = n.Destination + ".so"
		}

		n.Destination = filepath.Join(workdir, n.Destination) // Combine workdir and destination
	}

	if n.Type != "package" { // Binary or plugin
		runErr = os.MkdirAll(filepath.Dir(n.Destination), coreutils.NonGlobalFileMode)
	}

	if (n.Type == "package") && n.Source == "" { // If this is a package and source is not set
		n.Source = filepath.Join("src", n.SimpleName, "*.go") // Set our source to the package name
	}

	if runErr == nil { // If there wasn't any error creating the necessary directories
		args := []string{"build"}

		if n.Type != "package" { // Binary or plugin
			files := n.GetFiles("_test.go") // Exclude _test files

			if n.Type == "plugin" { // Plugin
				args = append(args, []string{"-buildmode", "plugin"}...)
			}

			args = append(args, []string{"-o", n.Destination}...)
			args = append(args, files...)
		} else { // Package
			args = append(args, n.SimpleName)
		}

		goCompilerOutput := coreutils.ExecCommand("go", args, true)

		if strings.Contains(goCompilerOutput, ".go") || strings.Contains(goCompilerOutput, "# ") { // If running the go build shows there are obvious issues
			runErr = errors.New(strings.TrimSpace(goCompilerOutput))
		} else { // If there was no obvious issues
			fmt.Println("Build successful.")

			if n.Type == "binary" {
				coreutils.ExecCommand("strip", []string{n.Destination}, true) // Strip the binary
			}

			sourceDir := filepath.Dir(n.Source)
			if goFiles, getErr := coreutils.GetFilesContains(sourceDir, ".go"); getErr == nil { // Get all files with .go extension
				if len(goFiles) != 0 { // If we managed to find files
					args := []string{"-s", "-w"}
					args = append(args, goFiles...)
					coreutils.ExecCommand("gofmt", args, false) // Run formatting
				}
			} else { // If we failed to get files
				runErr = errors.New("failed to get files from " + sourceDir + ": " + getErr.Error())
			}
		}
	} else { // If we failed to create the necessary directories
		fmt.Printf("failed to create the necessary directories:\n%s\n", runErr.Error())
		ToggleGoEnv(false)
	}

	return runErr
}

// ToggleGoEnv will toggle our usage of GOPATH and working directory
func ToggleGoEnv(on bool) error {
	var toggleEnvErr error
	if on {
		goDir := filepath.Join(workdir, "go")
		originalGoPath = os.Getenv("GOPATH") // Store the original GOPATH
		toggleEnvErr = os.Setenv("GOPATH", goDir)
		os.Chdir(goDir)
	} else {
		if toggleEnvErr = os.Setenv("GOPATH", originalGoPath); toggleEnvErr == nil { // If there was no error setting the environment
			toggleEnvErr = os.Chdir(workdir)
		}
	}

	return toggleEnvErr
}
