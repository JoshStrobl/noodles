package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"strings"
)

var project string              // Any project we're specifying for build
var noodlesCondensedName string // Condensed Noodles name
var workdir string              // Our working directory

// Commands

var rootCmd = &cobra.Command{
	Use:   "noodles",
	Short: "noodles is an opinionated manager for web apps.",
	Long: `noodles is an opinionated manager for web applications, enabling various functionality such as:
	- basic dependency management for built-in plugin support
	- compilation of project(s) in a configurable, ordered manner
	- configurable packing of project assets for distribution`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Use != "init" { // If we're not initializing
			ReadConfig() // Read the config

			if noodles.Name != "" { // If Name is set
				noodlesCondensedName = strings.ToLower(noodles.Name)                                          // Lowercase the noodles.Name
				noodlesCondensedName = strings.Replace(strings.TrimSpace(noodlesCondensedName), " ", "_", -1) // Trim the project name and replace any whitespace with _
			}
		}
	},
}

// Main

func init() {
	var getWdErr error
	workdir, getWdErr = os.Getwd() // Get the current working directory

	if getWdErr == nil {
		workdir = workdir + coreutils.Separator
	} else {
		fmt.Println("Failed to get the current working directory.")
		os.Exit(1)
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(scriptCmd)

	// Persistent Flags
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project to apply for specific commands")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
