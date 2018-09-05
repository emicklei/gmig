package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli"
)

// space is right after timestamp
const (
	logseparator = "~-------------- ------------------~"
	applied      = "--- applied ---"
	pending      = "... pending ..."
	execDo       = "...      do ..."
	execUndo     = "...    undo ..."
	stopped      = "... stopped ..."
)

func cmdCreateMigration(c *cli.Context) error {

	desc := c.Args().First()
	if len(desc) == 0 {
		printError("missing migration title")
		return errAbort
	}
	filename := NewFilename(desc)
	m := Migration{
		Description: desc,
		DoSection:   []string{"gcloud config list"},
		UndoSection: []string{"gcloud config list"},
	}
	yaml, err := m.ToYAML()
	if err != nil {
		printError("YAML creation failed")
		return errAbort
	}
	return ioutil.WriteFile(filename, []byte(yaml), os.FileMode(0644)) // -rw-r--r--, see http://permissions-calculator.org/
}

func cmdMigrationsUp(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	stopAfter := c.Args().Get(1) // empty if not specified
	all, err := LoadMigrationsBetweenAnd(mtx.migrationsPath, mtx.lastApplied, stopAfter)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	// if stopAfter is specified then it must be one of all
	found := false
	for _, each := range all {
		if stopAfter == each.Filename {
			found = true
			break
		}
	}
	// if lastApplied is after stopAfter then it is also not found but then we don't care
	if !found && stopAfter > mtx.lastApplied {
		reportError(mtx.stateProvider.Config(), "up until stop", errors.New("No such migration file: "+stopAfter))
		return errAbort
	}
	for _, each := range all {
		log.Println(logseparator)
		log.Println(execDo, pretty(each.Filename))
		if err := ExecuteAll(each.DoSection, mtx.config().shellEnv()); err != nil {
			reportError(mtx.stateProvider.Config(), "do", err)
			return errAbort
		}
		mtx.lastApplied = each.Filename
		// save after each succesful migration
		if err := mtx.stateProvider.SaveState(mtx.lastApplied); err != nil {
			reportError(mtx.stateProvider.Config(), "save state", err)
			return errAbort
		}
		// if not empty then stop after applying this migration
		if stopAfter == each.Filename {
			log.Println(stopped)
			log.Println(logseparator)
			break
		}
		log.Println(logseparator)
	}
	return nil
}

func cmdMigrationsDown(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	all, err := LoadMigrationsBetweenAnd(mtx.migrationsPath, "", mtx.lastApplied)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	lastMigration := all[len(all)-1]
	log.Println(logseparator)
	log.Println(execUndo, pretty(mtx.lastApplied))
	log.Println(logseparator)
	if err := ExecuteAll(lastMigration.UndoSection, mtx.config().shellEnv()); err != nil {
		reportError(mtx.stateProvider.Config(), "undo", err)
		return errAbort
	}
	// save after succesful migration
	previousFilename := ""
	if len(all) > 1 {
		previousFilename = all[len(all)-2].Filename
	}
	if err := mtx.stateProvider.SaveState(previousFilename); err != nil {
		reportError(mtx.stateProvider.Config(), "save state", err)
		return errAbort
	}
	return nil
}

func cmdMigrationsStatus(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	all, err := LoadMigrationsBetweenAnd(mtx.migrationsPath, "", "")
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	log.Println(logseparator)
	var last string
	prettyWidth := 0
	for _, each := range all {
		pf := pretty(each.Filename)
		if len(pf) > prettyWidth {
			prettyWidth = len(pf)
		}
	}
	for _, each := range all {
		status := applied
		if each.Filename > mtx.lastApplied {
			status = pending
			if len(last) > 0 && last != status {
				log.Println(logseparator)
			}
		}
		log.Printf("%s %-"+strconv.Itoa(prettyWidth)+"s (%s)\n", status, pretty(each.Filename), each.Filename)
		last = status
	}
	log.Println(logseparator)
	return nil
}

func cmdView(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	all, err := LoadMigrationsBetweenAnd(mtx.migrationsPath, "", "")
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	for _, each := range all {
		log.Println(logseparator)
		log.Println("View:", each.Filename)
		log.Println(logseparator)
		ExecuteAll(each.ViewSection, mtx.config().shellEnv())
	}
	return nil
}

func cmdInit(c *cli.Context) error {
	target := c.Args().First()
	if len(target) == 0 {
		printError("missing target name in command line")
		return errAbort
	}
	if err := os.MkdirAll(target, os.ModePerm|os.ModeDir); err != nil {
		printError(err.Error())
		return errAbort
	}
	location := filepath.Join(target, ConfigFilename)
	_, err := os.Stat(location)
	if err == nil {
		log.Println("config file [", location, "] already present.")
		cfg, err := LoadConfig(location)
		if err != nil {
			printError(err.Error())
			return errAbort
		}
		// TODO move to Config
		log.Println("config [ bucket=", cfg.Bucket, ",state=", cfg.LastMigrationObjectName, ",verbose=", cfg.verbose, "]")
		return nil
	}
	cfg := Config{
		LastMigrationObjectName: "gmig-last-migration",
		EnvironmentVars: map[string]string{
			"FOO": "bar",
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "\t")
	err = ioutil.WriteFile(location, data, os.ModePerm)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	return nil
}

func cmdExportProjectIAMPolicy(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	if err := ExportProjectsIAMPolicy(mtx.stateProvider.Config()); err != nil {
		printError(err.Error())
		return errAbort
	}
	return nil
}

func cmdExportStorageIAMPolicy(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	if err := ExportStorageIAMPolicy(mtx.stateProvider.Config()); err != nil {
		printError(err.Error())
		return errAbort
	}
	return nil
}
