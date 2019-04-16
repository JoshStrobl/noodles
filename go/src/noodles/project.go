package main

import (
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"strings"
)

// General NoodlesProject functions

// GetFiles will return all applicable files related to this project from Source
func (n *NoodlesProject) GetFiles() []string {
	var files []string

	if n.Source != "" { // If a source is defined
		fileName := filepath.Base(n.Source) // Get the file name

		if strings.HasPrefix(fileName, "*") { // If we're globbing
			files, _ = coreutils.GetFilesContains(n.SourceDir, filepath.Ext(fileName))
			if len(n.ExcludeItems) != 0 { // If we should be excluding items
				tmpFiles := []string{}

				for _, file := range files { // For each file
					if !ListContains(n.ExcludeItems, file) { // If the file should not be excluded
						tmpFiles = append(tmpFiles, file) // Add the file
					}
				}

				files = tmpFiles // Update
			}
		} else { // If we're not globbing
			files = []string{fileName} // Append the fileName to sources
		}
	}

	return files
}
