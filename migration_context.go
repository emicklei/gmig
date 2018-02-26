package main

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

type migrationContext struct {
	project       string
	lastApplied   string
	stateProvider StateProvider
}

func getMigrationContext(c *cli.Context) (ctx migrationContext, err error) {
	// allow target as folder name
	target := filepath.Base(c.Args().First())
	if len(target) == 0 {
		err = fmt.Errorf("missing target name in command line")
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
	if len(lastApplied) > 0 {
		e := checkExists(lastApplied)
		if e != nil {
			err = e
			return
		}
	}
	ctx.stateProvider = stateProvider
	ctx.lastApplied = lastApplied
	return
}

func (m migrationContext) shellEnv() (envs []string) {
	return append(envs, "PROJECT="+m.config().Project, "REGION="+m.config().Project, "ZONE="+m.config().Zone)
}

func (m migrationContext) config() Config {
	return m.stateProvider.Config()
}
