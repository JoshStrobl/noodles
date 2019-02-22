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

var temporaryTrackedCleanupFiles map[string][]string

func init() {
	temporaryTrackedCleanupFiles = make(map[string][]string)
}

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

// CleanupFiles will clean up any tracked files
func (p *GoPlugin) CleanupFiles(n *NoodlesProject) error {
	var cleanupErr error

	if cleanupFiles, exists := temporaryTrackedCleanupFiles[n.SimpleName]; exists { // If we have files to cleanup
		for _, fileName := range cleanupFiles { // For each file we need to cleanup
			if removeErr := os.Remove(filepath.Join(n.SourceDir, fileName)); removeErr != nil { // If we failed to remove this file
				cleanupErr = removeErr
				break
			}
		}

		temporaryTrackedCleanupFiles[n.SimpleName] = []string{} // Reset
	}

	return cleanupErr
}

// ConsolidateFiles will consolidate any files in child directories, if ConsolidateChildDirs is enabled
func (p *GoPlugin) ConsolidateFiles(n *NoodlesProject) error {
	var consolidateErr error

	if n.ConsolidateChildDirs { // If we should consolidate child directories into the root directory of the project
		if sourceDirFile, sourceOpenErr := os.Open(n.SourceDir); sourceOpenErr == nil { // If we successfully opened the directory to read the contents
			if files, fileReadDirErr := sourceDirFile.Readdir(-1); fileReadDirErr == nil { // If we successfully read the directory contents of the source dir
				for _, file := range files { // For each file
					if file.IsDir() { // Is this a child dir inside our source dir
						if copyErr := p.RecursiveCopy(filepath.Join(n.SourceDir, file.Name()), n); copyErr != nil { // If we failed to recursively copy the files
							consolidateErr = copyErr
							break
						}
					}
				}
			} else { // If we failed to read the contents of the source dir
				consolidateErr = sourceOpenErr
			}
		} else {
			consolidateErr = sourceOpenErr
		}
	}

	return consolidateErr
}

// Lint will lint our Go code. Lint takes a Noodles Project and the minimum acceptable confidence
func (p *GoPlugin) Lint(n *NoodlesProject, confidence float64) error {
	var lintErr error

	goFiles, getErr := coreutils.GetFilesContains(n.SourceDir, ".go") // Get all files with .go extension

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

	preRunErr = p.ConsolidateFiles(n)

	return preRunErr
}

// PostRun will reset our Go environment post-compilation
func (p *GoPlugin) PostRun(n *NoodlesProject) error {
	var postRunErr error

	postRunErr = p.CleanupFiles(n) // Cleanup any files related to this project
	postRunErr = ToggleGoEnv(false)

	return postRunErr
}

// RecursiveCopy will recursively copy files in child directories of the specific dir to the project source directory
// This function will rename the copied files to avoid conflicts. So if your file was x/y/z.go, it would change to x_y_z.go
func (p *GoPlugin) RecursiveCopy(dir string, n *NoodlesProject) error {
	var recursiveCopyErr error

	if dirFile, dirOpenErr := os.Open(dir); dirOpenErr == nil { // If we successfully opened the directory to read the contents
		if sourceDirItems, readDirErr := dirFile.Readdir(-1); readDirErr == nil { // Read all the contents of the directory as FileInfo structs
			for _, fileInfo := range sourceDirItems { // For each directory item
				if fileInfo.IsDir() { // If this is a directory
					if innerCopyErr := p.RecursiveCopy(filepath.Join(dir, fileInfo.Name()), n); innerCopyErr != nil { // Perform a recursive copy, if we fail to do so then...
						recursiveCopyErr = innerCopyErr
						break
					}
				} else { // If this is a file
					leadingPath := strings.Replace(dir, n.SourceDir, "", -1) // Remove the source directory
					originalFile := filepath.Join(dir, fileInfo.Name())
					conflictFreeFileName := strings.Replace(leadingPath, "/", "__", -1) + "__" + fileInfo.Name() // Replace all / with __ and add file name
					conflictFreePath := filepath.Join(n.SourceDir, conflictFreeFileName)

					if copyErr := coreutils.CopyFile(originalFile, conflictFreePath); copyErr == nil { // If we successfully copied the file to our SourceDir
						var cleanupFiles []string
						var exists bool

						if cleanupFiles, exists = temporaryTrackedCleanupFiles[n.SimpleName]; !exists { // Does not exist
							cleanupFiles = []string{} // Set as an empty slice
						}
						temporaryTrackedCleanupFiles[n.SimpleName] = append(cleanupFiles, conflictFreeFileName) // Add to cleanup files
					} else { // If we failed copying this file
						recursiveCopyErr = copyErr
						break
					}
				}
			}
		} else { // Failed to read the directory
			recursiveCopyErr = readDirErr
		}
	} else {
		recursiveCopyErr = dirOpenErr
	}

	return recursiveCopyErr
}

// RequiresPreRun is a stub function.
func (p *GoPlugin) RequiresPreRun(n *NoodlesProject) error {
	consolidateErr := p.ConsolidateFiles(n)
	os.Chdir(workdir)

	return consolidateErr
}

// RequiresPostRun is a stub function.
func (p *GoPlugin) RequiresPostRun(n *NoodlesProject) error {
	return p.CleanupFiles(n)
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
			goCompilerOutput = strings.TrimSpace(goCompilerOutput)              // Trim space
			goCompilerOutput = strings.Replace(goCompilerOutput, "__", "/", -1) // Replace all __ with /

			runErr = errors.New(goCompilerOutput)
		} else { // If there was no obvious issues
			fmt.Println("Build successful.")

			if n.Type == "binary" {
				coreutils.ExecCommand("strip", []string{n.Destination}, true) // Strip the binary
			}

			if goFiles, getErr := coreutils.GetFilesContains(n.SourceDir, ".go"); getErr == nil { // Get all files with .go extension
				if len(goFiles) != 0 { // If we managed to find files
					args := []string{"-s", "-w"}
					args = append(args, goFiles...)
					coreutils.ExecCommand("gofmt", args, false) // Run formatting
				}
			} else { // If we failed to get files
				runErr = errors.New("failed to get files from " + n.SourceDir + ": " + getErr.Error())
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
