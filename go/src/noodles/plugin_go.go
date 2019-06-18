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

var originalGoModule string
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
			filePath := p.GetCorrectGoFilePath(n, fileName)

			if removeErr := os.Remove(filePath); removeErr != nil { // If we failed to remove this file
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
					name := file.Name()
					if file.IsDir() && !ListContains(n.ExcludeItems, name) { // Is this a child dir inside our source dir and it isn't intentionally excluded
						if copyErr := p.RecursiveCopy(filepath.Join(n.SourceDir, name), n); copyErr != nil { // If we failed to recursively copy the files
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

// Format will run gofmt against the files in this project
func (p *GoPlugin) Format(n *NoodlesProject) error {
	var formatErr error

	if allFiles, getErr := coreutils.GetFilesContainsRecursive(n.SourceDir, ".go"); getErr == nil { // Get all files recursively
		if len(allFiles) > 0 {
			for _, file := range allFiles { // For each file
				file = strings.Replace(file, n.SourceDir, "", 1) // Remove any redundant source dir from path
				fullPath := p.GetCorrectGoFilePath(n, file)      // Get the right path for this file
				args := []string{"-s", "-w", fullPath}
				coreutils.ExecCommand("gofmt", args, false) // Run formatting
			}
		}
	} else { // If we failed to get files
		formatErr = errors.New("Failed to get files from " + n.SourceDir + ": " + getErr.Error())
	}

	return formatErr
}

// GetCorrectGoFilePath will attempt to get the correct path to the provided file
func (p *GoPlugin) GetCorrectGoFilePath(n *NoodlesProject, fileName string) string {
	var filePath string
	currentDir, _ := os.Getwd()

	if !strings.HasSuffix(currentDir, "/go") && !n.DisableNestedEnvironment { // If we're currently not in go and we don't have nested env disabled
		filePath = filepath.Join("go", n.SourceDir, fileName) // Ensure we add go before SourceDir and filename
	} else { // If we're currently in the go dir
		filePath = filepath.Join(n.SourceDir, fileName) // Only have SourceDir and filename
	}

	return filePath
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
									file := CleanupGoCompilerOutput(fileName)
									text := CleanupGoCompilerOutput(problem.LineText)

									// Example: test.go:24:34: test = errors.New("Hello world.")
									// error strings should not be capitalized or end with punctuation or a newline
									lineErr := strings.TrimSpace(text) // Trim any spacing
									fmt.Printf("%s:%d:%d: %s\n%s\n", file, problem.Position.Line, problem.Position.Column, lineErr, problem.Text)
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
func (p *GoPlugin) PreRun(n *NoodlesProject) (preRunErr error) {

	if !coreutils.ExecutableExists("go") { // If the go executable does not exist
		preRunErr = errors.New("go is not installed on your system. Please run noodles setup")
		return
	}

	if preRunErr = ToggleGoModules(n.EnableGoModules, false); preRunErr != nil { // Failed to set go modules support
		return
	}

	if !n.DisableNestedEnvironment { // If nested environment isn't disabled
		if preRunErr = ToggleGoEnv(true); preRunErr != nil { // Failed to toggle go environment
			return
		}
	}

	if preRunErr = p.Format(n); preRunErr != nil { // Failed to format the files in this project
		return
	}

	preRunErr = p.ConsolidateFiles(n) // Consolidate files

	return
}

// PostRun will reset our Go environment post-compilation
func (p *GoPlugin) PostRun(n *NoodlesProject) (postRunErr error) {
	if postRunErr = ToggleGoModules(false, true); postRunErr != nil { // Failed to reset go modules support
		return
	}

	postRunErr = p.CleanupFiles(n) // Cleanup any files related to this project

	if !n.DisableNestedEnvironment { // If we don't have nested environment disabled
		if postRunErr == nil { // If we successfully cleaned up the files and don't
			postRunErr = ToggleGoEnv(false)
		} else { // If we did not successfully clean up the files
			ToggleGoEnv(false) // Do not override our postRunErr from CleanupFiles, it is more important
		}
	}

	return
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
func (p *GoPlugin) RequiresPreRun(n *NoodlesProject) (requiresPreRunErr error) {
	requiresPreRunErr = p.ConsolidateFiles(n)
	os.Chdir(workdir)

	return
}

// RequiresPostRun is a stub function.
func (p *GoPlugin) RequiresPostRun(n *NoodlesProject) error {
	return p.CleanupFiles(n)
}

// Run will compile the provided project
func (p *GoPlugin) Run(n *NoodlesProject) error {
	var runErr error

	if n.Destination == "" { // If a destination is not set
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
		if !n.DisableNestedEnvironment { // If we haven't disabled nesting
			n.Source = filepath.Join("src", n.SimpleName, "*.go") // Set our source to the package name
		} else { // If we've disabled nesting, don't assume any directory
			n.Source = filepath.Join(n.SimpleName, "*.go") // Set our source to the simplename + *.go
		}
	}

	if runErr == nil { // If there wasn't any error creating the necessary directories
		args := []string{"build"}

		if n.Type != "package" { // Binary or plugin
			files := n.GetFiles() // Exclude _test files

			if n.Type == "plugin" { // Plugin
				args = append(args, []string{"-buildmode", "plugin"}...)
			}

			args = append(args, []string{"-o", n.Destination}...)
			args = append(args, files...)
		} else if n.Type == "package" && !n.DisableNestedEnvironment { // Package and we're using a nested env
			args = append(args, n.SimpleName) // Append the simple name of the package since that's what our GOPATH will recognize
		}

		goCompilerOutput := coreutils.ExecCommand("go", args, true)
		var buildSuccessful bool

		if len(goCompilerOutput) != 0 {
			compilerOutputLines := strings.Split(goCompilerOutput, "\n") // Split on newlines

			for _, compilerOutputLine := range compilerOutputLines { // For each line
				compilerOutputLine = strings.TrimSpace(compilerOutputLine) // Trim spacing

				if len(compilerOutputLine) != 0 &&
					(strings.Contains(compilerOutputLine, "can't") || // Typically can't load package, failure to import
						strings.Contains(compilerOutputLine, "cannot") || // Cannot load a package, namely when using Go modules with multiple paths in relative GOPATH
						strings.Contains(compilerOutputLine, ".go") ||
						strings.Contains(compilerOutputLine, "# ")) {
					buildSuccessful = false // Reset to false
					break                   // Break, with buildSuccessful being false
				} else {
					buildSuccessful = true // Temporarily accept build may be successful
				}
			}
		} else { // Absolutely no output, so a success (can happen when building and no errors, however not when pulling down modules at the same time)
			buildSuccessful = true
		}

		if buildSuccessful {
			fmt.Println("Build successful.")

			if n.Type == "binary" {
				coreutils.ExecCommand("strip", []string{n.Destination}, true) // Strip the binary
			}

			if debug {
				fmt.Println(goCompilerOutput) // Always output compiler content
			}
		} else {
			runErr = errors.New(CleanupGoCompilerOutput(goCompilerOutput))
		}
	} else { // If we failed to create the necessary directories
		fmt.Printf("failed to create the necessary directories:\n%s\n", runErr.Error())
	}

	return runErr
}

// CleanupGoCompilerOutput will handle the cleanup of any strings that would otherwise result from ConsolidateChildDirs
func CleanupGoCompilerOutput(output string) string {
	output = strings.TrimSpace(output)              // Trim space
	output = strings.Replace(output, "__", "/", -1) // Replace all __ with /
	return output
}

// ToggleGoEnv will toggle our usage of GOPATH and working directory
func ToggleGoEnv(on bool) (toggleEnvErr error) {
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

	return
}

// ToggleGoModules will toggle setting our GO111MODULE for go mod support
func ToggleGoModules(on bool, revert bool) (toggleModulesErr error) {
	if on {
		originalGoModule = os.Getenv("GO111MODULE")       // Store the original GO111MODULE
		toggleModulesErr = os.Setenv("GO111MODULE", "on") // Set on
	} else {
		if revert { // If we should just revert
			toggleModulesErr = os.Setenv("GO111MODULE", originalGoModule) // Set back to original
		} else {
			toggleModulesErr = os.Setenv("GO111MODULE", "off") // Set off
		}
	}

	return
}
