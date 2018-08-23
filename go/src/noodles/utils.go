// Misc. helpers utilities for noodles

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/stroblindustries/coreutils"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateHash will create a sha1sum of the provided bytes
func CreateHash(content []byte) string {
	hasher := sha1.New()
	hasher.Write(content)
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

// CopyFile will copy the source (file path) provided to the destination file
func CopyFile(source, destination string) error {
	var copyFileErr error
	destinationFolder := filepath.Dir(destination) // Get the folders leading up to the file

	if createDestFolderErr := os.MkdirAll(destinationFolder, 0755); createDestFolderErr != nil {
		copyFileErr = errors.New("failed to create " + destinationFolder + ": " + createDestFolderErr.Error())
	}

	if copyFileErr == nil {
		if destinationFile, createErr := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0755); createErr == nil { // Create a file to copy the contents into
			if sourceFile, openErr := os.OpenFile(source, os.O_RDONLY, 0755); openErr == nil {
				io.Copy(destinationFile, sourceFile) // Copy the contents
				sourceFile.Close()                   // Close project file
				destinationFile.Close()              // Close the temporary destination file
			} else {
				copyFileErr = errors.New("failed to open " + source + ": " + openErr.Error())
			}
		} else {
			copyFileErr = errors.New("failed to create " + destination + ": " + createErr.Error())
		}
	}

	return copyFileErr
}

// IsValidGitRemote will try to determine whether the URL provided is a valid git remote URL
func IsValidGitRemote(url string) bool {
	return strings.HasSuffix(url, ".git")
}

// IsYes will check if the provided input is yes
func IsYes(input string) bool {
	input = strings.ToLower(input)
	return (input == "y") || (input == "yes")
}

// ListContains will check if a string array contains a substring
func ListContains(list []string, substring string) bool {
	var contains bool

	for _, s := range list {
		if strings.Contains(s, substring) {
			contains = true
			break
		}
	}

	return contains
}

// PromptErrorCheck will check if we have a valid error from a prompt and if so, display and exit.
func PromptErrorCheck(promptErr error) {
	if promptErr != nil { // If we failed to get the prompt result
		fmt.Printf("failed to get the answer to our prompt: %s\n", promptErr.Error())
		os.Exit(1)
	}
}

// PromptExtensionValidate will check the provided input (provided via promptui) and return an error if a path does not contain a specific extension
func PromptExtensionValidate(expectedType, input string) error {
	var promptExtensionError error

	extension := filepath.Ext(input)                                                           // Get the extension
	projectExtension := strings.ToLower(strings.Replace(expectedType, "TypeScript", "ts", -1)) // Replace TypeScript with ts and ensure lowercase for Go and LESS

	if len(input) > 0 && extension != "" { // If we've provided input
		if extension[1:] != projectExtension { // If the extension provided by input (minus the prepended .) is not what we're expecting
			promptExtensionError = errors.New("source must be a specific " + expectedType + " file, or a glob (*." + projectExtension + ").")
		}
	}

	return promptExtensionError
}

// TextPromptValidate will get the requested input based on the message and validate it against our validate func
func TextPromptValidate(message string, validate validateFunc) string {
	var response string

	for {
		localResp := coreutils.InputMessage(message) // Get our input
		err := validate(localResp)

		if err == nil {
			response = localResp
			break
		} else {
			fmt.Println(err.Error())
		}
	}

	return response
}

// TextYNValidate will check if the provided input is a yes / no or y/n
func TextYNValidate(input string) error {
	var err error

	input = strings.ToLower(input)

	if (input != "yes") && (input != "y") && (input != "no") && (input != "n") {
		err = errors.New("must be a yes or no response")
	}

	return err
}
