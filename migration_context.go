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
	// actual absolute location of the configration file (gmig.yaml)
	configurationPath string
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
	ctx.configurationPath = fullPathToConfig
	ctx.migrationsPath = filepath.Dir(fullPathToConfig)
	// see if flag overrides this
	if migrationsHolder := c.String("migrations"); len(migrationsHolder) > 0 {
		newPath, perr := filepath.Abs(migrationsHolder)
		if ctx.config().verbose {
			log.Printf("override migrations path with [%s] from [%s] to [%s] err:[%v]\n", migrationsHolder, ctx.migrationsPath, newPath, perr)
		}
		if perr != nil {
			return
		}
		ctx.migrationsPath = newPath
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

func (m migrationContext) shellEnv() (envs []string) {
	envs = m.config().shellEnv()
	envs = append(envs, fmt.Sprintf("%s=%s", "GMIG_CONFIG_DIR", m.configurationPath))
	return
}
