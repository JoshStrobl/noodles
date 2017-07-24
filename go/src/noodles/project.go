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
		filePath := filepath.Dir(n.Source)
		fileName := filepath.Base(n.Source) // Get the file name
		
		if strings.HasPrefix(fileName, "*") { // If we're globbing
			files, _ = coreutils.GetFilesContains(filePath, filepath.Ext(fileName))
		} else { // If we're not globbing
			files = []string{fileName} // Append the fileName to sources
		}
	}

	return files
}