package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/emicklei/tre"
	"github.com/urfave/cli"
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
	return string(data), tre.New(err, "error reading state", "file", l.Configuration.LastMigrationObjectName)
}

// SaveState implements StateProvider
func (l FileStateProvider) SaveState(filename string) error {
	if l.Configuration.verbose {
		log.Println("writing local copy", l.Configuration.LastMigrationObjectName)
	}
	return ioutil.WriteFile(l.Configuration.LastMigrationObjectName, []byte(filename), os.ModePerm)
}

// Config implements StateProvider
func (l FileStateProvider) Config() Config {
	return l.Configuration
}

// DeleteState implements StateProvider
func (l FileStateProvider) DeleteState() {
	if l.Configuration.verbose {
		log.Println("deleting local copy", l.Configuration.LastMigrationObjectName)
	}
	os.Remove(l.Configuration.LastMigrationObjectName)
}

// read it once
var currentStateProvider StateProvider

func getStateProvider(c *cli.Context) (StateProvider, error) {
	if currentStateProvider != nil {
		return currentStateProvider, nil
	}
	verbose := c.GlobalBool("v")
	project := c.Args().First()
	// pre: project has been checked by the caller
	location := filepath.Join(project, ConfigFilename)
	if verbose {
		log.Println("loading configuration from", location)
	}
	cfg, err := LoadConfig(location)
	if err != nil {
		return currentStateProvider, tre.New(err, "error loading configuration (did you init?)")
	}
	cfg.verbose = cfg.verbose || verbose
	currentStateProvider = NewGCS(cfg)
	return currentStateProvider, nil
}

func checkExists(filename string) error {
	_, err := os.Stat(filename)
	return tre.New(err, "no such migration (wrong project?)", "file", filename)
}
