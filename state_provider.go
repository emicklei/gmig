package gmig

import (
	"io/ioutil"
	"log"
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
	data, err := ioutil.ReadFile(".gmig.state")
	return string(data), err
}
func (l FileStateProvider) SaveState(filename string) error {
	log.Println("[gmig] saving state")
	return ioutil.WriteFile(".gmig.state", []byte(filename), os.ModePerm)
}
