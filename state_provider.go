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
	if l.Configuration.verbose {
		log.Println("reading local copy", l.Configuration.LastMigrationObjectName)
	}
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

// for testing
var osRemove = os.Remove

// DeleteState implements StateProvider
func (l FileStateProvider) DeleteState() {
	if l.Configuration.verbose {
		log.Println("deleting local copy", l.Configuration.LastMigrationObjectName)
	}
	osRemove(l.Configuration.LastMigrationObjectName)
}

// read it once
var currentStateProvider StateProvider

func getStateProvider(c *cli.Context) (StateProvider, error) {
	if currentStateProvider != nil {
		return currentStateProvider, nil
	}
	verbose := c.GlobalBool("v")
	pathToConfig := c.Args().First()
	cfg, err := TryToLoadConfig(pathToConfig)
	if verbose {
		abs, _ := filepath.Abs(cfg.filename)
		log.Println("loading configuration from", abs)
	}
	if err != nil {
		workdir, _ := os.Getwd()
		abs, _ := filepath.Abs(cfg.filename)
		return currentStateProvider, tre.New(err, "error loading configuration (did you init?)", "path", pathToConfig, "workdir", workdir, "location", abs)
	}
	cfg.verbose = cfg.verbose || verbose
	currentStateProvider = NewGCS(*cfg)
	return currentStateProvider, nil
}
