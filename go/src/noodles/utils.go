// Misc. helpers utilities for noodles

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile will copy the source (file path) provided to the destination file
func CopyFile(source, destination string) error {
	destinationFolder := filepath.Dir(destination) // Get the folders leading up to the file

	if createDestFolderErr := os.MkdirAll(destinationFolder, 0755); createDestFolderErr != nil {
		return fmt.Errorf("Failed to create %s: %s\n", destinationFolder, createDestFolderErr.Error())
	}

	if destinationFile, createErr := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0755); createErr == nil { // Create a file to copy the contents into
		if sourceFile, openErr := os.OpenFile(source, os.O_RDONLY, 0755); openErr == nil {
			io.Copy(destinationFile, sourceFile) // Copy the contents
			sourceFile.Close()                   // Close project file
			destinationFile.Close()              // Close the temporary destination file

			return nil
		} else {
			return fmt.Errorf("Failed to open %s:\n\t%s\n", source, openErr.Error())
		}
	} else {
		return fmt.Errorf("Failed to create %s: %s\n", destination, createErr.Error())
	}
}
