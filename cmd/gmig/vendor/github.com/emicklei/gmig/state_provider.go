package gmig

import (
	"io/ioutil"
	"os"
)

// StateProvider knowns how to load state.
type StateProvider interface {
	// LoadState returns the last applied migration
	LoadState() (string, error)
	SaveState(filename string) error
}

// FileStateProvider use a local file to store state (last migration applied).
type FileStateProvider struct {
	Configuration Config
}

// LoadState implements StateProvider
func (l FileStateProvider) LoadState() (string, error) {
	data, err := ioutil.ReadFile(LastMigrationObjectName)
	return string(data), err
}

// SaveState implements StateProvider
func (l FileStateProvider) SaveState(filename string) error {
	return ioutil.WriteFile(LastMigrationObjectName, []byte(filename), os.ModePerm)
}
