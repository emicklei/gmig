package main

import (
	"errors"
	"fmt"
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
	skipped             = "--- skipped ---"
	skipping            = "... skipping .."
	conditionErrored    = "--- if error --"
	conditionError      = "... if error .."
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
	return os.WriteFile(filename, []byte(yaml), os.FileMode(0644)) // -rw-r--r--, see http://permissions-calculator.org/
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
	envs := mtx.shellEnv()
	// if lastApplied is after stopAfter then it is also not found but then we don't care
	if !found && stopAfter > mtx.lastApplied {
		reportError(mtx.stateProvider.Config(), envs, "up until stop", errors.New("No such migration file: "+stopAfter))
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
			if LogAll(each.IfExpression, each.DoSection, envs, true) != nil {
				reportError(mtx.stateProvider.Config(), envs, "plan do", err)
				return errAbort
			}
		} else {
			if err := ExecuteAll(each.IfExpression, each.DoSection, envs, c.GlobalBool("v")); err != nil {
				reportError(mtx.stateProvider.Config(), envs, "do", err)
				return errAbort
			}
			mtx.lastApplied = each.Filename
			// save after each succesful migration
			if err := mtx.stateProvider.SaveState(mtx.lastApplied); err != nil {
				reportError(mtx.stateProvider.Config(), envs, "save state", err)
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

func cmdMigrationsDownAll(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		return err
	}
	if mtx.lastApplied == "" {
		printWarning("There are no migrations to undo")
		return errAbort
	}
	//get all applied migrations
	all, err := LoadMigrationsBetweenAnd(".", "", mtx.lastApplied)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	for range all {
		err := cmdMigrationsDown(c)
		if err != nil {
			return err
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
	envs := mtx.shellEnv()
	if err := ExecuteAll(lastMigration.IfExpression, lastMigration.UndoSection, envs, c.GlobalBool("v")); err != nil {
		reportError(mtx.stateProvider.Config(), envs, "undo", err)
		return errAbort
	}
	// save after succesful migration
	previousFilename := ""
	if len(all) > 1 {
		previousFilename = all[len(all)-2].Filename
	}
	if err := mtx.stateProvider.SaveState(previousFilename); err != nil {
		reportError(mtx.stateProvider.Config(), envs, "save state", err)
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
	envs := mtx.shellEnv()
	for i, each := range all {
		var status string
		// check skipped
		pass, err := evaluateCondition(each.IfExpression, envs)
		isPending := each.Filename > mtx.lastApplied
		if err != nil {
			if isPending {
				status = conditionError
			} else {
				status = conditionErrored
			}
		} else {
			// no error condition
			if pass {
				if isPending {
					status = pending
				} else {
					status = applied
				}
			} else {
				if isPending {
					status = skipping
				} else {
					status = skipped
				}
			}
		}
		if err != nil {
			printWarning("if: expression is invalid:", err)
		}
		if i > 0 && isPending {
			log.Println(statusSeparator)
		}
		log.Printf("%s %s\n", status, pretty(each.Filename))
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
		if err := ExecuteAll(each.IfExpression, each.ViewSection, mtx.shellEnv(), c.GlobalBool("v")); err != nil {
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
