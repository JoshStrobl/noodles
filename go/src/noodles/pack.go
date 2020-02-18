package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
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
	if !coreutils.ExecutableExists("tar") { // Tar not on system
		trunk.LogFatal("tar does not exist on the system.")
	}

	os.RemoveAll(tmpDir) // Wipe our tmpDir

	var projectsToPack map[string]NoodlesProject

	if packProject == "" {
		trunk.LogInfo("Started packing.")
		projectsToPack = noodles.Projects
	} else {
		projectsToPack = map[string]NoodlesProject{
			packProject: noodles.Projects[packProject],
		}
	}

	if creationErr := os.Mkdir(tmpDir, 0755); creationErr != nil {
		trunk.LogErrRaw(fmt.Errorf("Failed to create our temporary directory:\n%s", creationErr.Error()))
		return
	}

	for projectName, project := range projectsToPack { // For each project
		trunk.LogInfo("Packing " + projectName)

		if project.Plugin != "" { // If a plugin is defined
			projectDestFolder := filepath.Dir(project.Destination)
			fileName := filepath.Base(project.Destination)
			fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			files := []string{fileName} // Have an array of files we should copy, at minimum the specified fileName

			if project.TarballLocation == "" { // If no tarball location
				trunk.LogWarn("No tarball location has been set for this project. We'll attempt to place this in a smart place.")

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
}

// TarContents will create a tar file out of the contents of our temporary directory and save it to the corresponding .tar file
func TarContents() {
	noodlesCondensedName := strings.ToLower(noodles.Name)                                         // Lowercase the workspace name
	noodlesCondensedName = strings.Replace(strings.TrimSpace(noodlesCondensedName), " ", "_", -1) // Trim whitespace and replace rest with _

	version := strconv.FormatFloat(noodles.Version, 'f', -1, 64) // Convert our float64 noodles.Version to a version string

	for _, compressor := range noodles.Distribution.TarCompressors { // For each compressor
		tarName := noodlesCondensedName + "-" + version + ".tar" // Create our initial tarball name

		switch compressor {
		case "bzip2": // bzip2 doesn't use .bzip2
			tarName += ".bz2" // Use .tar.bz2
		case "gzip": // gzip doesn't use .gzip
			tarName += ".gz" // Use .tar.gz
		case "lzma": // lzma
			tarName += ".lzma" // Use .tar.lzma
		case "xz": // xz
			tarName += ".xz" // Use .tar.xz
		case "zstd": // zstd doesn't use .zstd
			tarName += ".zst" // Use .tar.zst
		}

		tarArgs := []string{
			"-C",    // Change to the directory so we don't have the leading directory paths
			tmpDir,  // The directory we're compressing
			"-a",    // Auto compress based on archive suffix
			"-c",    // Create tar archive
			"-f",    // Specify file
			tarName, // Must specify tar name after -f
			".",     // Current directory, which is our tmpDir after changing
		}

		trunk.LogInfo("Creating " + tarName)
		coreutils.ExecCommand("tar", tarArgs, true)
	}
}
