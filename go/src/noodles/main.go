package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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
		ReadConfig() // Read the config (if it exists)
	},
}

// Main

var compileDocs bool

func init() {
	var getWdErr error
	workdir, getWdErr = os.Getwd() // Get the current working directory

	if getWdErr != nil { // If we failed to get the current working directory
		fmt.Printf("Failed to get the current working directory: %s", getWdErr.Error())
		os.Exit(1)
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(scriptCmd)
	rootCmd.Flags().BoolVarP(&compileDocs, "compile-docs", "c", false, "Compiles Noodle documentation. Strictly for noodles usage, not by projects using noodles.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	if compileDocs {
		docsRootPath := filepath.Join(workdir, "docs")
		manPagesPath := filepath.Join(docsRootPath, "man")
		mdPagesPath := filepath.Join(docsRootPath, "md")

		err := doc.GenMarkdownTree(rootCmd, mdPagesPath) // Generate Markdown files

		if err == nil {
			fmt.Println("Successfully generated Markdown pages.")
		} else {
			fmt.Printf("Failed to compile Markdown pages: %s\n", err.Error())
		}

		manHeader := &doc.GenManHeader{ // Create our header
			Title:   "noodles",
			Section: "1", // General commands
		}

		err = doc.GenManTree(rootCmd, manHeader, manPagesPath) // Generate Man pages

		if err == nil {
			fmt.Println("Successfully generated man pages.")
		} else {
			fmt.Printf("Failed to compile man pages: %s\n", err.Error())
		}
	}
}
