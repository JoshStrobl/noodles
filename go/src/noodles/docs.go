package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"path/filepath"
)

var genDocs = &cobra.Command{
	Use:     "gen-docs",
	Aliases: []string{"gd"},
	Short:   "Generates Noodles documentation",
	Long:    "Generates Noodles documentation",
	Run:     generateDocs,
	Hidden:  true,
}

// generateDocs will generate documentation
func generateDocs(cmd *cobra.Command, args []string) {
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
