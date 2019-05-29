package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/urfave/cli"
)

type migrationContext struct {
	// lastApplied is the filename of last migration, relative to migrationsPath
	lastApplied   string
	stateProvider StateProvider
	// folder that contains migrations files
	migrationsPath string
}

func getMigrationContext(c *cli.Context) (ctx migrationContext, err error) {
	pathToConfig := c.Args().First()
	if len(pathToConfig) == 0 {
		err = fmt.Errorf("missing path containing gmig.yaml in command line")
		return
	}
	stateProvider, err := getStateProvider(c)
	if err != nil {
		return
	}
	err = gcloudConfigSetProject(stateProvider.Config())
	if err != nil {
		return
	}
	lastApplied, err := stateProvider.LoadState()
	if err != nil {
		return
	}
	ctx.stateProvider = stateProvider
	fullPathToConfig, err := filepath.Abs(pathToConfig)
	if err != nil {
		return
	}
	ctx.migrationsPath = filepath.Dir(fullPathToConfig)
	// see if flag overrides this
	if migrationsHolder := c.String("migrations"); len(migrationsHolder) > 0 {
		ctx.migrationsPath, err = filepath.Abs(migrationsHolder)
		if err != nil {
			return
		}
	}
	if ctx.config().verbose {
		log.Println("reading migrations from", ctx.migrationsPath)
	}
	ctx.lastApplied = lastApplied
	if len(lastApplied) > 0 {
		e := checkExists(filepath.Join(ctx.migrationsPath, lastApplied))
		if e != nil {
			err = e
			return
		}
	}
	return
}

func (m migrationContext) config() Config {
	return m.stateProvider.Config()
}
