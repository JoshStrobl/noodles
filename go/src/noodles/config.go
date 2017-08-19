package main

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/stroblindustries/coreutils"
)

var noodles NoodlesConfig // Our Noodles Config

// ReadConfig will read any local noodles.toml that exists and returns an error or NoodlesConfig
func ReadConfig() error {
	_, convErr := toml.DecodeFile(workdir+"noodles.toml", &noodles)
	return convErr
}

// SaveConfig will save the NoodlesConfig to noodles.toml
func SaveConfig() error {
	var saveErr error
	buffer := new(bytes.Buffer) // Create a buffer for the encoder
	encoder := toml.NewEncoder(buffer) // Create a new toml encoder
	encoder.Indent = "\t" // Use a tab because we're opinionated

	if saveErr = encoder.Encode(noodles); saveErr == nil { // Encode our noodles struct into a buffer
		saveErr = coreutils.WriteOrUpdateFile("noodles.toml", buffer.Bytes(), coreutils.NonGlobalFileMode) // Write the noodles.toml as non-global
	}

	return saveErr
}
