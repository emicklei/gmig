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
type FileStateProvider struct{}

// LoadState implements StateProvider
func (l FileStateProvider) LoadState() (string, error) {
	data, err := ioutil.ReadFile(localStateFilename)
	return string(data), err
}

// SaveState implements StateProvider
func (l FileStateProvider) SaveState(filename string) error {
	return ioutil.WriteFile(localStateFilename, []byte(filename), os.ModePerm)
}
