package main

import (
	"archive/tar"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var tmpDir string

var packCmd = &cobra.Command{
	Use:               "pack",
	Short:             "Package configured assets for all or a specified project",
	Long:              "Package configured assets for all or a specified project into a distributable tarball",
	Run:               pack,
	DisableAutoGenTag: true,
}

var packProject string

func init() {
	tmpDir = filepath.Join(workdir, ".noodles-pack")
	packCmd.Flags().StringVarP(&packProject, "project", "p", "", "Name of a project we're packing")
}

// pack will package configured assets for a specified project into a tarball
func pack(cmd *cobra.Command, args []string) {
	var projectsToPack map[string]NoodlesProject

	if packProject == "" {
		fmt.Println("Packing all the things!")
		projectsToPack = noodles.Projects
	} else {
		projectsToPack = map[string]NoodlesProject{
			packProject: noodles.Projects[packProject],
		}
	}

	if creationErr := os.Mkdir(tmpDir, 0755); creationErr != nil {
		fmt.Printf("Failed to create our temporary directory:\n%s", creationErr.Error())
		return
	}

	for projectName, project := range projectsToPack { // For each project
		fmt.Println("Packing " + projectName)

		if project.Plugin != "" { // If a plugin is defined
			projectDestFolder := filepath.Dir(project.Destination)
			fileName := filepath.Base(project.Destination)
			fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			files := []string{fileName} // Have an array of files we should copy, at minimum the specified fileName

			if project.TarballLocation == "" { // If no tarball location
				fmt.Println("\tNo tarball location has been set for this project. We'll attempt to place this in a smart place.")

				switch project.Plugin {
				case "less":
					project.TarballLocation = "css/"
				case "typescript":
					project.TarballLocation = "js/"
					files = append(files, fileNameNoExt+".d.ts") // Add the definition file

					if project.Compress {
						files = append(files, fileNameNoExt+".min.js") // Add the minified file
					}
				}
			}

			for _, file := range files {
				CopyFile(filepath.Join(projectDestFolder, file), filepath.Join(tmpDir, project.TarballLocation, file)) // Copy this specific file
			}
		}
	}

	TarContents()
	os.RemoveAll(tmpDir) // Wipe our tmpDir
}

// TarContents will create a tar file out of the contents of our temporary directory and save it to the corresponding .tar file
func TarContents() {
	noodlesCondensedName := strings.ToLower(noodles.Name)                                         // Lowercase the workspace name
	noodlesCondensedName = strings.Replace(strings.TrimSpace(noodlesCondensedName), " ", "_", -1) // Trim whitespace and replace rest with _

	version := strconv.FormatFloat(noodles.Version, 'f', -1, 64) // Convert our float64 noodles.Version to a version string
	tarName := noodlesCondensedName + "-" + version + ".tar"
	file, createErr := os.Create(tarName) // Create a .tar file with the condensed name and version (ex. noodles-0.1.tar)

	if createErr == nil { // If we did not fail to create the .tar file
		tarWriter := tar.NewWriter(file)
		TarDirectory(tarWriter, tmpDir)
		tarWriter.Close() // Flush all contents to the file
		file.Close()      // Close the file

		tarContent, _ := ioutil.ReadFile(tarName)

		if len(tarContent) != 0 { // If there is content
			if coreutils.ExecutableExists("xz") { // If xz exists in PATH
				coreutils.ExecCommand("xz", []string{"-z", "-e", tarName}, true)
				os.Remove(tarName)
			} else {
				fmt.Println("xz does not exist on this system.")
			}
		} else {
			fmt.Println("No content found in tarball.")
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
				filePath := filepath.Join(directory, fileName)
				if file, fileOpenErr := os.Open(filePath); fileOpenErr == nil {
					fileStats, _ := file.Stat()
					fileHeader, fileHeaderErr := tar.FileInfoHeader(fileStats, "") // Create a new tar.Header

					if fileHeaderErr == nil { // If we didn't fail to create a FileInfoHeader
						if fileStats.IsDir() { // If this is a directory
							writer.WriteHeader(fileHeader) // Immediately write our fileHeader

							TarDirectory(writer, filePath)
						} else { // If this is a file
							relativeFolderName, _ := filepath.Rel(tmpDir, directory)

							if relativeFolderName != "." { // If we're not in the root directory of the tmpDir
								fileHeader.Name = filepath.Join(relativeFolderName, fileName)
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
