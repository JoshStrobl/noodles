package main

import (
	"errors"
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/stroblindustries/coreutils"
	xlint "golang.org/x/lint"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoPlugin is our Go plugin
type GoPlugin struct {
}

var originalGoModule string
var originalGoPath string
var originalGoPrivate string

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
func (p *GoPlugin) CleanupFiles(n *NoodlesProject) (cleanupErr error) {
	for key, cleanupFiles := range temporaryTrackedCleanupFiles { // For each key and list of files we should clean up
		for _, filePath := range cleanupFiles { // For each file we need to cleanup
			if cleanupErr = os.Remove(filePath); cleanupErr != nil { // If we failed to remove this file
				break
			}
		}

		temporaryTrackedCleanupFiles[key] = []string{} // Reset
	}

	return
}

// ConsolidateFiles will consolidate any files in child directories, if ConsolidateChildDirs is enabled
func (p *GoPlugin) ConsolidateFiles(n *NoodlesProject) (consolidateErr error) {
	if n.ConsolidateChildDirs { // If we should consolidate child directories into the root directory of the project
		p.Flatten(n.SimpleName, n.SourceDir, n.SourceDir, n.ExcludeItems) // Ensure all child directories within the root of our project are flattened
	}

	if n.EnableGoModules && !n.DisableNestedEnvironment { // If we've enabled Go Modules and have not disabled the use of our nested environment
		coreutils.ExecCommand("go", []string{"mod", "download"}, false) // Ensure we've pre-cached the modules before changing them
		pkgModPath := filepath.Join(workdir, "go", "pkg", "mod")        // Set up our mod path

		var nestedNoodleWorkspacesFilesList []string
		if nestedNoodleWorkspacesFilesList, consolidateErr = coreutils.GetFilesContainsRecursive(pkgModPath, "noodles.toml"); consolidateErr != nil { // Check for any directory which has noodles.toml
			return
		}

		for _, namespaceDir := range nestedNoodleWorkspacesFilesList { // For each reference to noodles.toml
			dir := filepath.Dir(namespaceDir) // Ensure we remove noodles.toml from path

			chmodErr := filepath.Walk(dir, func(dirEntryName string, info os.FileInfo, err error) error {
				return os.Chmod(dirEntryName, 0770)
			})

			if chmodErr != nil { // Ensure it is actually writable
				fmt.Printf("Failed to change permission: %s\n", chmodErr)
			}

			repoName := filepath.Base(dir) // Get the repo name

			if flattenErr := p.Flatten(repoName, dir, dir, n.ExcludeItems); flattenErr != nil { // Flatten this Noodles Workspace
				fmt.Printf("Failed to flatten %s: %s\n", flattenErr)
			}
		}
	}

	return
}

// Flatten will attempt to get all files from the srcDir (or nested) and consolidate them in the targetDir specified
// This will also automatically add them to our "tracked" files to clean up
func (p *GoPlugin) Flatten(tempTrackingKey string, srcDir string, targetDir string, exclude []string) (flattenErr error) {
	var sourceDirFile *os.File
	if sourceDirFile, flattenErr = os.Open(srcDir); flattenErr != nil {
		return
	}

	var sourceDirItems []os.FileInfo
	if sourceDirItems, flattenErr = sourceDirFile.Readdir(-1); flattenErr != nil { // Read all the items from this directory
		return
	}

	for _, fileInfo := range sourceDirItems { // For each directory item
		fileName := fileInfo.Name()
		originalFilePath := filepath.Join(srcDir, fileName)

		if ListContains(exclude, fileName) { // If this item should be excluded
			continue // This actually skips this specific iteration and moves onto the next
		}

		if fileInfo.IsDir() && !strings.HasPrefix(fileName, ".") { // If this is a non-hidden directory
			nestedSrcDir := filepath.Join(srcDir, fileName)

			if innerFlattenErr := p.Flatten(tempTrackingKey, nestedSrcDir, targetDir, exclude); innerFlattenErr != nil { // Perform a recursive flatten
				flattenErr = innerFlattenErr
				break
			}
		} else if !fileInfo.IsDir() { // If this is a file
			leadingPath := strings.Replace(srcDir, targetDir, "", -1) // Ensure we remove all references to targetDir from srcDir path
			leadingPath = strings.TrimPrefix(leadingPath, "/")        // Trim any starting /

			if originalFilePath != filepath.Join(targetDir, fileName) { // If this isn't at the root of targetDir (if sourceDir and targetDir are same)
				conflictFreeFileName := strings.Replace(leadingPath, "/", "__", -1) + "__" + fileName // Replace all / with __ and add file name
				conflictFreePath := filepath.Join(targetDir, conflictFreeFileName)

				if copyErr := coreutils.CopyFile(originalFilePath, conflictFreePath); copyErr == nil { // If we successfully copied the file to our target directory
					var cleanupFiles []string
					var exists bool

					if cleanupFiles, exists = temporaryTrackedCleanupFiles[tempTrackingKey]; !exists { // Does not exist
						cleanupFiles = []string{} // Set as an empty slice
					}

					temporaryTrackedCleanupFiles[tempTrackingKey] = append(cleanupFiles, conflictFreePath) // Add to cleanup files
				} else { // If we failed copying this file
					flattenErr = copyErr
					break
				}
			}
		}
	}

	return
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

// ModInit will ensure our Go Modules is initted if we don't have a go.mod already
func (p *GoPlugin) ModInit(n *NoodlesProject) {
	if !n.EnableGoModules { // Go Modules not enabled
		return
	}

	modFile, modOpenErr := os.Open("go.mod")
	defer modFile.Close()

	if modOpenErr != nil { // Failed to open file
		if os.IsPermission(modOpenErr) { // Permission error opening file
			trunk.LogErrRaw(modOpenErr)
		} else if os.IsNotExist(modOpenErr) { // File doesn't exist
			coreutils.ExecCommand("go", []string{"mod", "init"}, false) // Run go mod init
		}
	}
}

// Tidy will tidy up Go Modules
func (p *GoPlugin) ModTidy(n *NoodlesProject) {
	if !n.EnableGoModules {
		trunk.LogErr(n.SimpleName + " does not have Go Modules Enabled")
		return
	}

	coreutils.ExecCommand("go", []string{"mod", "tidy"}, false) // Run go mod tidy to remove unused deps
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

	p.ModInit(n) // Mod Init if necessary

	if !n.DisableNestedEnvironment { // If nested environment isn't disabled
		if preRunErr = ToggleGoEnv(true); preRunErr != nil { // Failed to toggle go environment
			return
		}
	}

	if len(n.Private) != 0 { // Have Private / non-public module URIs set
		originalGoPrivate = os.Getenv("GOPRIVATE")                       // Get the original GOPRIVATE value
		preRunErr = os.Setenv("GOPRIVATE", strings.Join(n.Private, ",")) // Set GOPRIVATE to comma-separated n.Private list

		if preRunErr != nil {
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

	if originalGoPrivate != "" { // Non-empty original value
		os.Setenv("GOPRIVATE", originalGoPrivate)
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
func (p *GoPlugin) Run(n *NoodlesProject) (runErr error) {
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
		if runErr = os.MkdirAll(filepath.Dir(n.Destination), coreutils.NonGlobalFileMode); runErr != nil { // Failed to create directories
			runErr = fmt.Errorf("failed to create the necessary directories:\n%s\n", runErr.Error())
			return
		}
	}

	if (n.Type == "package") && n.Source == "" { // If this is a package and source is not set
		if !n.DisableNestedEnvironment { // If we haven't disabled nesting
			n.Source = filepath.Join("src", n.SimpleName, "*.go") // Set our source to the package name
		} else { // If we've disabled nesting, don't assume any directory
			n.Source = filepath.Join(n.SimpleName, "*.go") // Set our source to the simplename + *.go
		}
	}

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

	builder := exec.Command("go", args...) // Create an os/exec command for go building

	stderr, pipeErr := builder.StderrPipe()
	stdout, outErr := builder.StdoutPipe()

	if pipeErr != nil {
		runErr = errors.New("Failed to create pipe for stderr for build command")
	} else if outErr != nil {
		runErr = errors.New("Failed to create pipe for stdout for build command")
	}

	if runErr != nil { // Failed during pipe creation
		return
	}

	defer stderr.Close()
	defer stdout.Close()

	builder.Start() // Start, use instead of Run for piping

	stderrOutput, _ := ioutil.ReadAll(stderr) // Read from stderr
	stdoutOutput, _ := ioutil.ReadAll(stdout) // Read from stdout

	if len(stderrOutput) != 0 { // Have stderr content
		return errors.New(CleanupGoCompilerOutput(string(stderrOutput[:]))) // Return an error that is our stderr content
	} else if len(stdoutOutput) != 0 && debug { // Have stdout content and debugging
		trunk.LogDebug(CleanupGoCompilerOutput(string(stdoutOutput[:])))
	}

	if n.Type == "binary" {
		coreutils.ExecCommand("strip", []string{n.Destination}, true) // Strip the binary
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
