package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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
		if cmd.Use != "new" || (cmd.Use == "new" && (newProjectName != "") || (newScriptName != "")) { // If we're not creating a new Noodles workspace
			if conf, readErr := ReadConfig(filepath.Join(workdir, "noodles.toml")); readErr == nil { // Read the config
				noodles = conf
			} else {
				trunk.LogFatal(fmt.Sprintf("noodles.toml appears to have the following issue(s):\n%s\n", readErr.Error()))
			}
		}
	},
	DisableAutoGenTag: true,
}

// Main

func init() {
	var getWdErr error
	workdir, getWdErr = os.Getwd() // Get the current working directory

	if getWdErr != nil { // If we failed to get the current working directory
		trunk.LogFatal("Failed to get the current working directory: " + getWdErr.Error())
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(genDocs)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(scriptCmd)
	rootCmd.AddCommand(tidyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
