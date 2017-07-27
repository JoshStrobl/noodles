package main

import (
	"archive/tar"
	"fmt"
	"github.com/solus-project/xzed"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	var projectsToPack map[string]NoodlesProject

	if project == "" {
		fmt.Println("Packing all the things!")
		projectsToPack = noodles.Projects
	} else {
		projectsToPack = map[string]NoodlesProject{
			project: noodles.Projects[project],
		}
	}

	now := strconv.FormatInt(time.Now().Unix(), 10) // Convert the current Unix time to a string
	tmpDir = os.TempDir() + coreutils.Separator + "noodles-" + now + coreutils.Separator

	for projectName, project := range projectsToPack { // For each project
		fmt.Println("Packing " + projectName)

		if project.Plugin != "" { // If a plugin is defined
			switch project.Plugin {
			case "go":
				if project.Binary { // If we're making a binary, copy it
					relativePathToBuild, _ := filepath.Rel("build"+coreutils.Separator, project.Destination)
					folders := filepath.Dir(project.Destination)
					binaryName := filepath.Base(project.Destination)

					if binaryName == relativePathToBuild { // If the binary is directly in build folder
						coreutils.CopyFile(project.Destination, tmpDir+binaryName) // Copy the file
					} else { // If is in an inner folder
						childDirectoriesOfFolder := strings.TrimPrefix(folders, "build"+coreutils.Separator)
						coreutils.CopyDirectory(folders, tmpDir+childDirectoriesOfFolder) // Copy the github.com/ulikunitz/xzdirectory instead
					}
				}
			}
		}
	}

	TarContents()
	os.RemoveAll(tmpDir) // Wipe our tmpDir
}

// TarContents will create a tar file out of the contents of our temporary directory and save it to the cooresponding .tar file
func TarContents() {
	version := strconv.FormatFloat(noodles.Version, 'f', -1, 64) // Convert our float64 noodles.Version to a version string
	tarName := noodlesCondensedName + "-" + version + ".tar"
	file, createErr := os.Create(tarName) // Create a .tar file with the condensed name and version (ex. noodles-0.1.tar)

	if createErr == nil { // If we did not fail to create the .tar file
		tarWriter := tar.NewWriter(file)
		TarDirectory(tarWriter, tmpDir)
		tarWriter.Close() // Flush all contents to the file
		file.Close()      // Close the file

		if xzfile, xzCreateErr := os.Create(tarName + ".xz"); xzCreateErr == nil { // Create an xz file
			tarContent, _ := ioutil.ReadFile(tarName)

			if len(tarContent) != 0 { // If there is content
				if xzWriter, xzWriterErr := xzed.NewWriterLevel(xzfile, xzed.BestCompression); xzWriterErr == nil {
					xzWriter.Write(tarContent)
					xzWriter.Close()
					xzfile.Close()
					os.Remove(tarName) // Remove the .tar file since it is no longer needed.
				} else {
					fmt.Println("Failed to create a compressed tarball.")
				}
			} else {
				fmt.Println("No content found in tarball.")
			}
		}
	} else {
		fmt.Println("Failed to create our .tar file.")
	}
}

// TarDirectory will take our tar Writer and a directory and write its contents
func TarDirectory(writer *tar.Writer, directory string) {
	if dir, openErr := os.Open(directory); openErr == nil { // Create an os.File struct via Open
		if dirContents, readErr := dir.Readdirnames(-1); readErr == nil { // If there was no readErr
			for _, fileName := range dirContents { // For each FileInfo struct in dirContents
				if file, fileOpenErr := os.Open(directory + coreutils.Separator + fileName); fileOpenErr == nil {
					fileStats, _ := file.Stat()
					fileHeader, fileHeaderErr := tar.FileInfoHeader(fileStats, "") // Create a new tar.Header

					if fileHeaderErr == nil { // If we didn't fail to create a FileInfoHeader
						if fileStats.IsDir() { // If this is a directory
							writer.WriteHeader(fileHeader) // Immediately write our fileHeader

							TarDirectory(writer, directory+fileName)
						} else { // If this is a file
							relativeFolderName, _ := filepath.Rel(tmpDir, directory)

							if relativeFolderName != "." { // If we're not in the root directory of the tmpDir
								fileHeader.Name = relativeFolderName + coreutils.Separator + fileName
							}

							writer.WriteHeader(fileHeader) // Immediately write our fileHeader

							bytes := make([]byte, fileStats.Size()) // Make a bytes array the size of the file
							file.Read(bytes)                        // Read into bytes
							writer.Write(bytes)                     // Write the bytes into the tar.Writer
						}
					} else {
						fmt.Printf("Failed to create a FileInfoHeader:\n%s\n", fileHeaderErr)
					}
				} else { // If we failed to open the file
					fmt.Println("Failed to open "+fileName+":\n%s\n", fileOpenErr)
				}
			}
		} else { // If we failed to read the directory
			fmt.Printf("Failed to read %s\n", directory)
		}
	} else { // If we failed to open the directory
		fmt.Printf("Failed to open %s\n", directory)
	}
}
