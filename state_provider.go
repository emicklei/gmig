package main

import (
	"io/ioutil"
	"log"
	"os"
)

// StateProvider knowns how to load state.
type StateProvider interface {
	// LoadState returns the last applied migration
	LoadState() (string, error)
	SaveState(filename string) error
	Config() Config
}

// FileStateProvider use a local file to store state (last migration applied).
type FileStateProvider struct {
	Configuration Config
}

// LoadState implements StateProvider
func (l FileStateProvider) LoadState() (string, error) {
	data, err := ioutil.ReadFile(l.Configuration.LastMigrationObjectName)
	return string(data), err
}

// SaveState implements StateProvider
func (l FileStateProvider) SaveState(filename string) error {
	return ioutil.WriteFile(l.Configuration.LastMigrationObjectName, []byte(filename), os.ModePerm)
}

// Config implements StateProvider
func (l FileStateProvider) Config() Config {
	return l.Configuration
}

// DeleteState implements StateProvider
func (l FileStateProvider) DeleteState() {
	if l.Configuration.Verbose {
		log.Println("deleting local copy", l.Configuration.LastMigrationObjectName)
	}
	os.Remove(l.Configuration.LastMigrationObjectName)
}
