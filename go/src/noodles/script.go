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
	Use:     "script",
	Aliases: []string{"run-script"},
	Short:   "Run a custom script",
	Long:    "Run a custom script",
	Run:     script,
}

var verbose bool

func init() {
	scriptCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose mode.")
}

func script(cmd *cobra.Command, args []string) {
	if len(args) != 0 { // If no argument are passed
		for _, arg := range args { // For each argument
			if _, exists := noodles.Scripts[arg]; exists { // If this
				RunScript(arg)
			} else {
				fmt.Printf("%s is not a valid script. Exiting.\n", arg)
				break
			}
		}
	} else {
		if len(noodles.Scripts) != 0 { // If there are scripts set
			fmt.Printf("No scripts passed to run. Here is a list:\n\n")
			for name, script := range noodles.Scripts {
				line := name

				if script.Description != "" {
					line += ": " + script.Description
				}

				fmt.Println("-", line)
			}
			fmt.Printf("\n") // Intentional padding ad end of output
		} else {
			fmt.Println("No scripts defined in your noodles.toml file.")
		}
	}
}

// RunScript will run the script provided
func RunScript(name string) {
	script, _ := noodles.Scripts[name] // Get our script

	if script.Exec != "" { // If there is an executable
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

		output := coreutils.ExecCommand(script.Exec, script.Arguments, script.Redirect)

		if (script.File != "") && script.Redirect { // If we should redirect output to a file
			coreutils.WriteOrUpdateFile(script.File, []byte(output), coreutils.NonGlobalFileMode)
		}

		os.Chdir(workdir) // Change back to the work dir if needed
	} else {
		fmt.Printf("No executable set for the script: %s", name)
	}
}