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
	tempDir       string
}

// for testing
var osTempDir = os.TempDir

func NewFileStateProvider(c Config) FileStateProvider {
	return FileStateProvider{
		Configuration: c,
		tempDir:       osTempDir(),
	}
}

func (l FileStateProvider) stateFilename() string {
	return filepath.Join(l.tempDir, l.Configuration.LastMigrationObjectName)
}

// LoadState implements StateProvider
func (l FileStateProvider) LoadState() (string, error) {
	if l.Configuration.verbose {
		d, _ := os.Getwd()
		log.Println("reading local copy", l.stateFilename(), ",cwd=", d)
	}
	data, err := os.ReadFile(l.stateFilename())
	return string(data), tre.New(err, "error reading state", "tempDir", l.tempDir, "lastMigration", l.Configuration.LastMigrationObjectName)
}

// SaveState implements StateProvider
func (l FileStateProvider) SaveState(filename string) error {
	if l.Configuration.verbose {
		d, _ := os.Getwd()
		log.Println("writing local copy", l.stateFilename(), ",cwd=", d)
	}
	return ioutil.WriteFile(l.stateFilename(), []byte(filename), os.ModePerm)
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
		d, _ := os.Getwd()
		log.Println("deleting local copy", l.stateFilename(), ",cwd=", d)
	}
	osRemove(l.stateFilename())
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
	if verbose && err == nil {
		abs, _ := filepath.Abs(cfg.filename)
		log.Println("loading configuration from", abs)
	}
	if err != nil {
		workdir, _ := os.Getwd()
		abs := "?"
		if cfg != nil {
			abs, _ = filepath.Abs(cfg.filename)
		}
		return currentStateProvider, tre.New(err, "error loading configuration (did you init?)", "path", pathToConfig, "workdir", workdir, "location", abs)
	}
	cfg.verbose = cfg.verbose || verbose
	currentStateProvider = NewGCS(*cfg)
	return currentStateProvider, nil
}
