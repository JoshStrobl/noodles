package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/pelletier/go-toml"
	"github.com/stroblindustries/coreutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// NoodlesConfig is the configuration of global properties of Noodles.
type NoodlesConfig struct {
	Description  string
	Distribution *NoodlesDistributionConfig
	License      string
	Name         string
	Projects     map[string]NoodlesProject
	Scripts      map[string]NoodlesScript
	Version      float64
}

// NoodlesDistributionConfig is the configuration for distribution
type NoodlesDistributionConfig struct {
	TarCompressors []string
}

var SupportedTarCompressors []string // SupportedTarCompressions are various compressions we officially support
var noodles NoodlesConfig            // Our Noodles Config
var desiredConfigPath string

func init() {
	SupportedTarCompressors = []string{"bzip2", "gzip", "lzma", "xz", "zstd"}
}

// ReadConfig will read any local noodles.toml that exists and returns an error or NoodlesConfig
func ReadConfig(configPath string) (conf NoodlesConfig, readConfigErr error) {
	desiredConfigPath = configPath
	var configContent []byte

	if configContent, readConfigErr = ioutil.ReadFile(configPath); readConfigErr != nil { // If we failed to read the config
		if os.IsNotExist(readConfigErr) { // If the file does not exist
			readConfigErr = errors.New("noodles.toml does not exist in this directory")
		} else if os.IsPermission(readConfigErr) { // If we don't have the necessary permissions to read this file
			readConfigErr = errors.New("noodles.toml is not readable in the provided path")
		}

		return
	}

	if readConfigErr = toml.Unmarshal(configContent, &conf); readConfigErr != nil { // If we failed to unmarshal our config content bytes into the conf NoodlesConfig
		return
	}

	if conf.Distribution == nil { // Distribution not set
		conf.Distribution = &NoodlesDistributionConfig{
			TarCompressors: []string{"zstd"},
		}
	}

	compressors := conf.Distribution.TarCompressors

	if len(compressors) == 0 {
		compressors = []string{"zstd"} // Default to zstd
		conf.Distribution.TarCompressors = compressors
	}

	for _, compressor := range compressors {
		if !ListContains(SupportedTarCompressors, compressor) { // If the provided compressor isn't supported
			readConfigErr = errors.New("Must use a supported compressor: " + strings.Join(SupportedTarCompressors, ","))
			return
		}
	}

	for name, project := range conf.Projects { // For each noodles project
		if project.ConsolidateChildDirs && (project.SimpleName == "") { // No SimpleName defined, and it'll be required during consolidation
			project.SimpleName = name
		}

		project.SourceDir = filepath.Dir(project.Source)

		if project.SourceDir != "" { // If SourceDir has content
			project.SourceDir = project.SourceDir + "/" // Add trailing /
		}

		if project.Type == "go" { // If this is a Go project
			if len(project.ExcludeItems) == 0 { // No items
				project.ExcludeItems = []string{"pkg/", "_test.go"} // Add pkg folder and _test.go files
			} else { // Has items
				if !ListContains(project.ExcludeItems, "pkg") {
					project.ExcludeItems = append(project.ExcludeItems, "pkg")
				}

				if !ListContains(project.ExcludeItems, "_test.go") {
					project.ExcludeItems = append(project.ExcludeItems, "_test.go")
				}
			}
		}

		conf.Projects[name] = project
	}

	return
}

// SaveConfig will save the NoodlesConfig to noodles.toml
func SaveConfig() (saveErr error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	encoder := toml.NewEncoder(writer).Indentation("\t") // Create a new toml encoder and set its indentation to \t

	if saveErr = encoder.Encode(noodles); saveErr == nil { // Encode our noodles struct into a buffer
		saveErr = coreutils.WriteOrUpdateFile(desiredConfigPath, buffer.Bytes(), coreutils.NonGlobalFileMode) // Write the noodles.toml as non-global
	}

	return saveErr
}
