package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
)

var project string // Any project we're specifying for build
var workdir string // Our working directory

// Commands

var rootCmd = &cobra.Command{
	Use: "noodles",
	Short: "noodles is an opinionated manager for web apps.",
	Long: `noodles is an opinionated manager for web applications, enabling various functionality such as:
	- basic dependency management for built-in plugin support
	- compilation of project(s) in a configurable, ordered manner
	- configurable packing of project assets for distribution`,
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

	// Persistent Flags
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project to apply for specific commands")
}

func main() {
	if err :=rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}