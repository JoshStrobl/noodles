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

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize noodles",
	Long: "Initialize noodles by generating a basic YAML configuration file",
	Run: initNoodles,
}

var lintCmd = &cobra.Command {
	Use: "lint",
	Short: "Validates the existing noodles.yml",
	Long: "Validates the existing noodles.yml",
	Run: lint,
}

var packCmd = &cobra.Command{
	Use: "pack",
	Short: "Package configured assets for all or a specified project",
	Long: "Package configured assets for all or a specified project into a distributable tarball",
	Run: pack,
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