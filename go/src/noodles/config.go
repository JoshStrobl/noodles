package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/stroblindustries/coreutils"
	"path/filepath"
	"regexp"
	"strings"
)

// NoodlesConfig is the configuration of global properties of Noodles.
type NoodlesConfig struct {
	Description string
	License     string
	Name        string
	Projects    map[string]NoodlesProject
	Scripts     map[string]NoodlesScript
	Version     float64
}

var noodles NoodlesConfig // Our Noodles Config

// ReadConfig will read any local noodles.toml that exists and returns an error or NoodlesConfig
func ReadConfig(configPath string) (conf NoodlesConfig, readConfigErr error) {
	if _, convErr := toml.DecodeFile(configPath, &conf); convErr == nil { // Decode our config
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
	} else { // If there was an error decoding
		if strings.Contains(convErr.Error(), "no such file or directory") {
			readConfigErr = errors.New("noodles.toml does not exist in this directory")
		} else { // If this is some sort of other error, sanitize it and return a new convErr
			sanitizedErrMessage := strings.Replace(convErr.Error(), "unmarshal", "convert", -1) // Change "unmarshal" to a human language
			sanitizedErrMessage = strings.Replace(sanitizedErrMessage, "!!", "", -1)
			sanitizedErrMessage = strings.Replace(sanitizedErrMessage, "`", "", -1) // Remove any ` wrapping types

			re := regexp.MustCompile(`line\s\d+:\s[\s\S]+$`)                                    // Only get line N: message
			lineErrors := re.FindAllString(sanitizedErrMessage, -1)                             // Find all strings
			sanitizedErrMessage = strings.Replace(strings.Join(lineErrors, "\n"), "  ", "", -1) // Join all with newline and remove unnecessary whitespace
			readConfigErr = errors.New(sanitizedErrMessage)                                     // Create a sanitized error
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
