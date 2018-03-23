package main

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli"
)

func cmdMigrationsSetState(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printWarning(err.Error())
	}
	filename := c.Args().Get(1) // 0=path, 1=relative filename
	if !c.GlobalBool("q") {     // be quiet
		if !promptForYes(fmt.Sprintf("Are you sure to overwrite the current last applied migration [%s -> %s] (y/N)? ", mtx.lastApplied, filename)) {
			return errAbort
		}
	}
	if err := checkExists(filepath.Join(mtx.migrationsPath, filename)); err != nil {
		printError(err.Error())
		return errAbort
	}
	if err := mtx.stateProvider.SaveState(filename); err != nil {
		printError(err.Error())
		return errAbort
	}
	return nil
}

func cmdRundoOnly(c *cli.Context) error {
	return runSectionOnly(c, true)
}

func cmdRunUndoOnly(c *cli.Context) error {
	return runSectionOnly(c, false)
}

func runSectionOnly(c *cli.Context, isDo bool) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printWarning(err.Error())
	}
	section := "do"
	if !isDo {
		section = "undo"
	}
	filename := c.Args().Get(1) // 0=path, 1=relative filename
	if !c.GlobalBool("q") {     // be quiet
		if !promptForYes(fmt.Sprintf("Are you sure to run the [%s] section of migration [%s] (y/N)? ", section, filename)) {
			return errAbort
		}
	}
	full := filepath.Join(mtx.migrationsPath, filename)
	if err := checkExists(full); err != nil {
		printError(err.Error())
		return errAbort
	}
	m, err := LoadMigration(full)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	lines := m.DoSection
	if !isDo {
		lines = m.UndoSection
	}
	if err := ExecuteAll(lines, mtx.config().shellEnv()); err != nil {
		reportError(mtx.stateProvider.Config(), section, err)
		return errAbort
	}
	return nil
}
