package main

// Script Functionality

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strings"
)

var scriptCmd = &cobra.Command{
	Use:               "script",
	Aliases:           []string{"run-script"},
	Short:             "Run a custom script",
	Long:              "Run a custom script",
	Run:               script,
	DisableAutoGenTag: true,
}

var verbose bool
var selectedScript string

func init() {
	scriptCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose mode.")
	scriptCmd.Flags().StringVarP(&selectedScript, "script", "s", "", "Name of the script we're running")
}

func script(cmd *cobra.Command, args []string) {
	if selectedScript == "" { // If no script is set
		for name := range noodles.Scripts {
			RunScript(name)
		}
	} else { // If a script is set
		RunScript(selectedScript)
	}
}

// RunScript will run the script provided
func RunScript(name string) {
	script, _ := noodles.Scripts[name] // Get our script

	if script.Exec != "" { // If there is an executable
		fmt.Printf("Running script: %s\n", name)

		if script.Directory != "" { // If we should run this command in a directory
			failedToChange := os.Chdir(filepath.Join(workdir, script.Directory)) // Change to the directory

			if failedToChange != nil { // If we failed to change to the directory
				fmt.Printf("Failed to change to the following directory: %s\n", script.Directory)

				if verbose {
					fmt.Printf("Full error: %s\n", failedToChange)
				}

				return // Don't continue with exec
			}
		}

		if verbose {
			commandRunning := script.Exec + " " + (strings.Join(script.Arguments, " "))
			fmt.Printf("Running: %s\n", commandRunning)
		}

		if script.UseGoEnv { // If we should be enforcing Go env
			ToggleGoEnv(true) // Toggle env on
		}

		output := coreutils.ExecCommand(script.Exec, script.Arguments, script.Redirect)

		if (script.File != "") && script.Redirect { // If we should redirect output to a file
			coreutils.WriteOrUpdateFile(script.File, []byte(output), coreutils.NonGlobalFileMode)
		}

		if script.UseGoEnv { // If we should be enforcing Go env
			ToggleGoEnv(false) // Toggle env off
		}

		os.Chdir(workdir) // Change back to the work dir if needed
	} else {
		fmt.Printf("No executable set for the script: %s", name)
	}
}
