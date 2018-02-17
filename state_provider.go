package gmig

import (
	"io/ioutil"
	"os"
)

// StateProvider knowns how to load state.
type StateProvider interface {
	// LoadState returns the last applied migration
	LoadState() (string, error)
	SaveState() error
}

type FileStateProvider struct{}

func (l FileStateProvider) LoadState() (string, error) {
	data, err := ioutil.ReadFile(localStateFilename)
	return string(data), err
}
func (l FileStateProvider) SaveState(filename string) error {
	//log.Println("[gmig] saving state")
	return ioutil.WriteFile(localStateFilename, []byte(filename), os.ModePerm)
}
