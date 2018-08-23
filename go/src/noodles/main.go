package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// Plugins
var goPlugin GoPlugin
var lessPlugin LessPlugin
var typescriptPlugin TypeScriptPlugin

var workdir string // Our working directory

// Commands

var rootCmd = &cobra.Command{
	Use:   "noodles",
	Short: "noodles is an opinionated manager for web apps.",
	Long: `noodles is an opinionated manager for web applications, enabling various functionality such as:
	- basic dependency management for built-in plugin support
	- compilation of project(s) in a configurable, ordered manner
	- configurable packing of project assets for distribution`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ReadConfig() // Read the config (if it exists)
	},
}

// Main

func init() {
	var getWdErr error
	workdir, getWdErr = os.Getwd() // Get the current working directory

	if getWdErr != nil { // If we failed to get the current working directory
		fmt.Printf("Failed to get the current working directory: %s", getWdErr.Error())
		os.Exit(1)
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(genDocs)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(scriptCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
