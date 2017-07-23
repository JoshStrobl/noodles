package main

import (
	"github.com/spf13/cobra"
	"os"
)

var noodles NoodlesConfig // Our Noodles Config
var project string // Any project we're specifying for build

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

var packCmd = &cobra.Command{
	Use: "pack",
	Short: "Package configured assets for all or a specified project",
	Long: "Package configured assets for all or a specified project into a distributable tarball",
	Run: pack,
}

// Main

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(packCmd)

	// Persistent Flags
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project to apply for specific commands")
}

func main() {
	if err :=rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}