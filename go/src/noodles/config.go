package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/stroblindustries/coreutils"
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

func init() {
	SupportedTarCompressors = []string{"bzip2", "gzip", "lzma", "xz", "zstd"}
}

// ReadConfig will read any local noodles.toml that exists and returns an error or NoodlesConfig
func ReadConfig(configPath string) (conf NoodlesConfig, readConfigErr error) {
	if _, convErr := toml.DecodeFile(configPath, &conf); convErr == nil { // Decode our config
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
	} else { // If there was an error decoding'
		if strings.Contains(convErr.Error(), "no such file or directory") {
			readConfigErr = errors.New("noodles.toml does not exist in this directory")
		} else { // If this is some sort of other error, sanitize it and return a new convErr
			sanitizedErrMessage := strings.Replace(convErr.Error(), "unmarshal", "convert", -1) // Change "unmarshal" to a human language
			sanitizedErrMessage = strings.Replace(sanitizedErrMessage, "!!", "", -1)
			sanitizedErrMessage = strings.Replace(sanitizedErrMessage, "`", "", -1) // Remove any ` wrapping types
			readConfigErr = errors.New(sanitizedErrMessage)                         // Create a sanitized error
		}
	}

	return
}

// SaveConfig will save the NoodlesConfig to noodles.toml
func SaveConfig() error {
	var saveErr error
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	encoder := toml.NewEncoder(writer) // Create a new toml encoder
	encoder.Indent = "\t"              // Use a tab because we're opinionated

	if saveErr = encoder.Encode(noodles); saveErr == nil { // Encode our noodles struct into a buffer
		saveErr = coreutils.WriteOrUpdateFile("noodles.toml", buffer.Bytes(), coreutils.NonGlobalFileMode) // Write the noodles.toml as non-global
	}

	return saveErr
}
