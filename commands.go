package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

// space is right after timestamp
const (
	statusSeparator     = "~-------------- ------------------~"
	viewSeparatorTop    = "~---------------------------------------------------------------~"
	viewSeparatorBottom = " --------------------------------------------------------------- "
	applied             = "--- applied ---"
	pending             = "... pending ..."
	execDo              = "...      do ..."
	execUndo            = "...    undo ..."
	stopped             = "... stopped ..."
)

func cmdCreateMigration(c *cli.Context) error {

	desc := c.Args().First()
	if len(desc) == 0 {
		printError("missing migration title")
		return errAbort
	}
	filename := NewFilename(desc)
	defaultCommands := []string{"gcloud config list"}
	doSection, undoSection, viewSection := defaultCommands, defaultCommands, defaultCommands
	if doValue := c.String("do"); len(doValue) > 0 {
		doSection = strings.Split(doValue, "\n")
	}
	if undoValue := c.String("undo"); len(undoValue) > 0 {
		undoSection = strings.Split(undoValue, "\n")
	}
	if viewValue := c.String("view"); len(viewValue) > 0 {
		viewSection = strings.Split(viewValue, "\n")
	}
	m := Migration{
		Description: desc,
		Filename:    filename,
		DoSection:   doSection,
		UndoSection: undoSection,
		ViewSection: viewSection,
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
		log.Println(statusSeparator)
		log.Println(execDo, pretty(each.Filename))
		if err := ExecuteAll(each.DoSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
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
			log.Println(statusSeparator)
			break
		}
		log.Println(statusSeparator)
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
	log.Println(statusSeparator)
	log.Println(execUndo, pretty(mtx.lastApplied))
	log.Println(statusSeparator)
	if err := ExecuteAll(lastMigration.UndoSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
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
	log.Println(statusSeparator)
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
				log.Println(statusSeparator)
			}
		}
		log.Printf("%s %-"+strconv.Itoa(prettyWidth)+"s (%s)\n", status, pretty(each.Filename), each.Filename)
		last = status
	}
	log.Println(statusSeparator)
	return nil
}

func cmdView(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	var all []Migration
	if len(c.Args()) == 2 {
		localMigrationFilename := filepath.Base(c.Args().Get(1))
		if len(localMigrationFilename) > 0 {
			one, err := LoadMigration(filepath.Join(mtx.migrationsPath, localMigrationFilename))
			if err != nil {
				printError(err.Error())
				return errAbort
			}
			all = append(all, one)
		}
	} else {
		all, err = LoadMigrationsBetweenAnd(mtx.migrationsPath, "", "")
		if err != nil {
			printError(err.Error())
			return errAbort
		}
	}
	for _, each := range all {
		log.Println(viewSeparatorTop)
		log.Printf(" %s (%s)\n", pretty(each.Filename), each.Filename)
		log.Println(viewSeparatorBottom)
		if each.Filename > mtx.lastApplied {
			log.Println(" ** this migration is pending...")
			break
		}
		if len(each.ViewSection) == 0 {
			log.Println(" ** this migration has no commands to describe its change on infrastructure.")
		}
		if mtx.config().verbose {
			log.Printf("executing view section (%d commands)\n", len(each.ViewSection))
		}
		if err := ExecuteAll(each.ViewSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
			printError(err.Error())
			return errAbort
		}
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
