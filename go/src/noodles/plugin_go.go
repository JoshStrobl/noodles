package main

import (
	"fmt"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strings"
)

var originalGoPath string

// This is the Go plugin code

func (n *NoodlesProject) Go(project string) {
	if !coreutils.ExecutableExists("go") { // If the go executable exists
		fmt.Println("Go is not installed on your system. Please run noodles setup.")
		return
	}

	if n.Destination == "" { // If no destination is set
		n.Destination = "build" + coreutils.Separator + project // Set destination to build/name (as binary)
	}

	if createDirsErr := os.MkdirAll(filepath.Dir(n.Destination), coreutils.NonGlobalFileMode); createDirsErr != nil { // Make all the necessary directories we need to
		fmt.Printf("Failed to create the necessary directories:\n%s\n", createDirsErr.Error())
		return
	}

	args := []string{"-o", n.Destination}

	originalGoPath = os.Getenv("GOPATH") // Store the original GOPATH
	os.Setenv("GOPATH", workdir + "go")

	os.Chdir(workdir + coreutils.Separator + "go") // Change to our go directory
	args = append(args, n.Source)

	goCompilerOutput := coreutils.ExecCommand("go", args, false)

	if strings.Contains(goCompilerOutput, ".go") || strings.Contains(goCompilerOutput, "# command") { // If running the go build shows there are obvious issues
		fmt.Println(goCompilerOutput)
	} else { // If there was no obvious issues
		fmt.Println("Build successful.")
	}

	os.Setenv("GOPATH", originalGoPath)
	os.Chdir(workdir)
}