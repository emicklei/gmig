package main

import (
	"errors"
	"fmt"
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
	execPlan            = "...    plan ..."
	stopped             = "... stopped ..."
)

func cmdCreateMigration(c *cli.Context) error {

	desc := c.Args().First()
	if len(desc) == 0 {
		printError("missing migration title")
		return errAbort
	}
	filename := NewFilenameWithIndex(desc)
	defaultCommands := []string{"gcloud config list"}
	doSection, undoSection, viewSection := defaultCommands, defaultCommands, []string{}
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
	return runMigrations(c, !true)
}

func cmdMigrationsPlan(c *cli.Context) error {
	return runMigrations(c, true)
}

func runMigrations(c *cli.Context, isLogOnly bool) error {
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
	prettyWidth := largestWidthOf(all)
	for _, each := range all {
		log.Println(statusSeparator)
		leadingTitle := execDo
		if isLogOnly {
			leadingTitle = execPlan
		}
		log.Printf("%s %-"+strconv.Itoa(prettyWidth)+"s (%s)\n", leadingTitle, pretty(each.Filename), each.Filename)
		if isLogOnly {
			log.Println("")
			if LogAll(each.DoSection, mtx.config().shellEnv(), true); err != nil {
				reportError(mtx.stateProvider.Config(), "plan do", err)
				return errAbort
			}
		} else {
			if err := ExecuteAll(each.IfExpression, each.DoSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
				reportError(mtx.stateProvider.Config(), "do", err)
				return errAbort
			}
			mtx.lastApplied = each.Filename
			// save after each succesful migration
			if err := mtx.stateProvider.SaveState(mtx.lastApplied); err != nil {
				reportError(mtx.stateProvider.Config(), "save state", err)
				return errAbort
			}
		}
		// if not empty then stop after applying this migration
		if stopAfter == each.Filename {
			log.Println(stopped)
			log.Println(statusSeparator)
			break
		}
	}
	return nil
}

func cmdMigrationsDown(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	if mtx.lastApplied == "" {
		printWarning("There are no migrations to undo")
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
	if err := ExecuteAll(lastMigration.IfExpression, lastMigration.UndoSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
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

func largestWidthOf(list []Migration) int {
	prettyWidth := 0
	for _, each := range list {
		pf := pretty(each.Filename)
		if len(pf) > prettyWidth {
			prettyWidth = len(pf)
		}
	}
	return prettyWidth
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
	for _, each := range all {
		status := applied
		if each.Filename > mtx.lastApplied {
			status = pending
			if len(last) > 0 && last != status {
				log.Println(statusSeparator)
			}
		}
		log.Printf("%s %s\n", status, pretty(each.Filename))
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
		if err := ExecuteAll(each.IfExpression, each.ViewSection, mtx.config().shellEnv(), c.GlobalBool("v")); err != nil {
			printError(err.Error())
			return errAbort
		}
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

func cmdExportEnv(c *cli.Context) error {
	pathToConfig := c.Args().First()
	config, err := TryToLoadConfig(pathToConfig)
	if err != nil {
		printError(err.Error())
		return errAbort
	}

	tmpl := "export %s=%s\n"
	fmt.Printf(tmpl, "PROJECT", config.Project)
	fmt.Printf(tmpl, "REGION", config.Region)
	fmt.Printf(tmpl, "ZONE", config.Zone)
	for key, value := range config.EnvironmentVars {
		fmt.Printf(tmpl, key, value)
	}
	return nil
}
