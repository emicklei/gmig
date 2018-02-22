package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// space is right after timestamp
const logseparator = "~-------------- --------------~"

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
	return ioutil.WriteFile(filename, []byte(yaml), os.ModePerm)
}

type migrationContext struct {
	project       string
	lastApplied   string
	stateProvider StateProvider
}

func getMigrationContext(c *cli.Context) (ctx migrationContext, err error) {
	// allow project as folder name
	project := filepath.Base(c.Args().First())
	if len(project) == 0 {
		err = fmt.Errorf("missing project name in command line")
		return
	}
	stateProvider, err := getStateProvider(c)
	if err != nil {
		return
	}
	err = gcloudConfigSetProject(stateProvider.Config(), project)
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
	ctx.project = project
	ctx.lastApplied = lastApplied
	return
}

func cmdMigrationsUp(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	all, err := LoadMigrationsBetweenAnd(mtx.lastApplied, c.Args().First())
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	for _, each := range all {
		log.Println(logseparator)
		log.Println(each.Filename)
		log.Println(logseparator)
		if err := ExecuteAll(each.DoSection, []string{"PROJECT=" + mtx.project}); err != nil {
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
	return nil
}

func cmdMigrationsDown(c *cli.Context) error {
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	all, err := LoadMigrationsBetweenAnd("", mtx.lastApplied)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	lastMigration := all[len(all)-1]
	log.Println(logseparator)
	log.Println(mtx.lastApplied)
	log.Println(logseparator)
	if err := ExecuteAll(lastMigration.UndoSection, []string{"PROJECT=" + mtx.project}); err != nil {
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
	all, err := LoadMigrationsBetweenAnd("", "")
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	log.Println(logseparator)
	var last string
	for _, each := range all {
		status := "--- applied ---"
		if each.Filename > mtx.lastApplied {
			status = "... pending ..."
			if len(last) > 0 && last != status {
				log.Println(logseparator)
			}
		}
		log.Println(status, each.Filename)
		last = status
	}
	log.Println(logseparator)
	return nil
}

func cmdInit(c *cli.Context) error {
	project := c.Args().First()
	if len(project) == 0 {
		printError("missing project name in command line")
		return errAbort
	}
	if err := os.MkdirAll(project, os.ModePerm|os.ModeDir); err != nil {
		printError(err.Error())
		return errAbort
	}
	location := filepath.Join(project, ConfigFilename)
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
		Bucket:                  "your-accessible-bucket",
		LastMigrationObjectName: "gmig-last-migration",
	}
	data, _ := json.Marshal(cfg)
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
	filename, err := ExportProjectsIAMPolicy(mtx.stateProvider.Config(), mtx.project)
	if err != nil {
		printError(err.Error())
		return err
	}
	if c.Bool("s") {
		if err := mtx.stateProvider.SaveState(filename); err != nil {
			printError(err.Error())
			return err
		}
	}
	return nil
}
