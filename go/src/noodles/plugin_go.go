package main

import (
	"fmt"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strings"
)

// This is the Go plugin

var originalGoPath string

func (n *NoodlesProject) Go(project string) {
	if !coreutils.ExecutableExists("go") { // If the go executable exists
		fmt.Println("Go is not installed on your system. Please run noodles setup.")
		return
	}

	ToggleGoEnv(true)

	os.Chdir(workdir + coreutils.Separator + "go") // Change to our go directory

	if n.Destination == "" { // If no destination is set
		n.Destination = "build" + coreutils.Separator + project // Set destination to build/name (as binary)
	}

	n.Destination = workdir + n.Destination

	if createDirsErr := os.MkdirAll(filepath.Dir(n.Destination), coreutils.NonGlobalFileMode); createDirsErr != nil { // Make all the necessary directories we need to
		fmt.Printf("Failed to create the necessary directories:\n%s\n", createDirsErr.Error())
		ToggleGoEnv(false)
		return
	}

	files := n.GetFiles()
	args := []string{"build", "-o", n.Destination}
	args = append(args, files...)

	goCompilerOutput := coreutils.ExecCommand("go", args, true)

	if strings.Contains(goCompilerOutput, ".go") || strings.Contains(goCompilerOutput, "# command") { // If running the go build shows there are obvious issues
		fmt.Println(strings.TrimSpace(goCompilerOutput))
	} else { // If there was no obvious issues
		fmt.Println("Build successful.")
	}

	ToggleGoEnv(false)
}

// ToggleGoEnv will toggle our usage of GOPATH and working directory
func ToggleGoEnv(on bool) {
	if on {
		originalGoPath = os.Getenv("GOPATH") // Store the original GOPATH
		os.Setenv("GOPATH", workdir+"go")
	} else {
		os.Setenv("GOPATH", originalGoPath)
		os.Chdir(workdir)
	}
}
