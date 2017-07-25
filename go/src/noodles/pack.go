package main

import (
	"archive/tar"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var tmpDir string

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Package configured assets for all or a specified project",
	Long:  "Package configured assets for all or a specified project into a distributable tarball",
	Run:   pack,
}

// pack will package configured assets for a specified project into a tarball
func pack(cmd *cobra.Command, args []string) {
	if project == "" {
		fmt.Println("Packing all the things!")
	} else {
		fmt.Printf("Packing %s\n", project)
	}

	now := strconv.FormatInt(time.Now().Unix(), 10) // Convert the current Unix time to a string
	tmpDir = os.TempDir() + coreutils.Separator + "noodles-" + now + coreutils.Separator

	for _, project := range noodles.Projects { // For each project
		if project.Plugin != "" { // If a plugin is defined
			switch (project.Plugin) {
				case "go":
					if project.Binary { // If we're making a binary, copy it
						binName := filepath.Base(project.Destination)

						if copyErr := coreutils.CopyFile(project.Destination, tmpDir + binName); copyErr != nil {
							fmt.Println(copyErr)
						}
					}
			}
		}
	}

	TarContents()
}

// TarContents will create a tar file out of the contents of our temporary directory and save it to the cooresponding .tar file
func TarContents() {
	version := strconv.FormatFloat(noodles.Version, 'f', -1, 64) // Convert our float64 noodles.Version to a version string
	file, createErr := os.Create(noodlesCondensedName + "-" + version + ".tar") // Create a .tar file with the condensed name and version (ex. noodles-0.1.tar)

	if createErr == nil { // If we did not fail to create the .tar file
		tarWriter := tar.NewWriter(file)
		TarDirectory(tarWriter, tmpDir)
		tarWriter.Close() // Flush all contents to the file
		file.Close() // Close the file
	} else {
		fmt.Println("Failed to create our .tar file.")
	}

	os.RemoveAll(tmpDir) // Wipe our tmpDir
}

// TarDirectory will take our tar Writer and a directory and write its contents
func TarDirectory(writer *tar.Writer, directory string) {
	fmt.Println("Packaging " + directory)
	if dir, openErr := os.Open(directory); openErr == nil { // Create an os.File struct via Open
		if dirContents, readErr := dir.Readdirnames(-1); readErr == nil { // If there was no readErr
			for _, fileName := range dirContents { // For each FileInfo struct in dirContents
				if file, fileOpenErr := os.Open(directory + coreutils.Separator + fileName); fileOpenErr == nil {
					fileStats, _ := file.Stat()
					fileHeader, fileHeaderErr := tar.FileInfoHeader(fileStats, "") // Create a new tar.Header

					if fileHeaderErr == nil { // If we didn't fail to create a FileInfoHeader
						writer.WriteHeader(fileHeader) // Write our fileHeader

						if fileStats.IsDir() { // If this is a directory
							TarDirectory(writer, directory + coreutils.Separator + fileName)
						} else { // If this is a file
							bytes := make([]byte, fileStats.Size()) // Make a bytes array the size of the file
							file.Read(bytes) // Read into bytes
							writer.Write(bytes) // Write the bytes into the tar.Writer
						}
					} else {
						fmt.Printf("Failed to create a FileInfoHeader:\n%s\n", fileHeaderErr)
					}
				} else { // If we failed to open the file
					fmt.Println("Failed to open " + fileName + ":\n%s\n", fileOpenErr)
				}
			}
		} else { // If we failed to read the directory
			fmt.Printf("Failed to read %s\n", directory)
		}
	} else { // If we failed to open the directory
		fmt.Printf("Failed to open %s\n", directory)
	}
}