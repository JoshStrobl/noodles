package main

import (
	"errors"
	"github.com/stroblindustries/coreutils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var noodles NoodlesConfig // Our Noodles Config

// Read will read any local noodles.yml that exists and returns an error or NoodlesConfig
func (n NoodlesConfig) Read() error {
	var configBytes []byte
	var readErr error

	if configBytes, readErr = ioutil.ReadFile("noodles.yml"); readErr == nil { // Read the contents of noodles.yml
		if len(configBytes) != 0 { // If the file isn't empty
			readErr = yaml.Unmarshal(configBytes, &n)
		} else {
			readErr = errors.New("noodles.yml is empty. Please init a noodles project.")
		}
	}

	return readErr
}

// Save will save the NoodlesConfig to noodles.yml
func (n NoodlesConfig) Save() error {
	var config []byte
	var saveErr error

	if config, saveErr = yaml.Marshal(&n); saveErr == nil { // Marshal our project NoodlesConfig
		saveErr = coreutils.WriteOrUpdateFile("noodles.yml", config, coreutils.NonGlobalFileMode) // Write the noodles.yml as non-global
	}

	return saveErr
}