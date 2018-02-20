package main

import (
	"encoding/json"
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
	filename := NewFilename(desc)
	m := Migration{
		Description: desc,
		DoSection:   []string{"gcloud config list"},
		UndoSection: []string{"gcloud config list"},
	}
	yaml, err := m.ToYAML()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(yaml), os.ModePerm)
}

func cmdMigrationsUp(c *cli.Context) error {
	project := c.Args().First()
	gcloudConfigSetProject(project)
	stateProvider := getStateProvider(c)
	lastApplied, err := stateProvider.LoadState()
	if err != nil {
		return err
	}
	if len(lastApplied) > 0 {
		err := checkExists(lastApplied)
		if err != nil {
			return err
		}
	}
	all, err := LoadMigrationsBetweenAnd(lastApplied, c.Args().First())
	if err != nil {
		return err
	}
	for _, each := range all {
		log.Println(logseparator)
		log.Println(each.Filename)
		log.Println(logseparator)
		if err := ExecuteAll(each.DoSection, []string{"PROJECT=" + project}); err != nil {
			reportError(stateProvider.Config(), "do", err)
			return err
		}
		lastApplied = each.Filename
		// save after each succesful migration
		if err := stateProvider.SaveState(lastApplied); err != nil {
			reportError(stateProvider.Config(), "save state", err)
			return err
		}
	}
	return nil
}

func cmdMigrationsDown(c *cli.Context) error {
	project := c.Args().First()
	gcloudConfigSetProject(project)
	stateProvider := getStateProvider(c)
	lastApplied, err := stateProvider.LoadState()
	if err != nil {
		reportError(stateProvider.Config(), "load state", err)
		return err
	}
	if len(lastApplied) > 0 {
		err := checkExists(lastApplied)
		if err != nil {
			return err
		}
	}
	all, err := LoadMigrationsBetweenAnd("", lastApplied)
	if err != nil {
		return err
	}
	lastMigration := all[len(all)-1]
	log.Println(logseparator)
	log.Println(lastApplied)
	log.Println(logseparator)
	if err := ExecuteAll(lastMigration.UndoSection, []string{"PROJECT=" + project}); err != nil {
		reportError(stateProvider.Config(), "undo", err)
		return err
	}
	// save after succesful migration
	previousFilename := ""
	if len(all) > 1 {
		previousFilename = all[len(all)-2].Filename
	}
	if err := stateProvider.SaveState(previousFilename); err != nil {
		reportError(stateProvider.Config(), "save state", err)
		return err
	}
	return nil
}

func cmdMigrationsStatus(c *cli.Context) error {
	stateProvider := getStateProvider(c)
	lastApplied, err := stateProvider.LoadState()
	if err != nil {
		reportError(stateProvider.Config(), "load state", err)
		return err
	}
	if len(lastApplied) > 0 {
		err := checkExists(lastApplied)
		if err != nil {
			return err
		}
	}
	all, err := LoadMigrationsBetweenAnd("", "")
	if err != nil {
		return err
	}
	log.Println(logseparator)
	var last string
	for _, each := range all {
		status := "--- applied ---"
		if each.Filename > lastApplied {
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
	if err := os.MkdirAll(project, os.ModePerm|os.ModeDir); err != nil {
		return err
	}
	location := filepath.Join(project, ConfigFilename)
	_, err := os.Stat(location)
	if err == nil {
		log.Println("config file [", location, "] already present.")
		cfg, err := LoadConfig(location)
		if err != nil {
			log.Println("cannot read configuration", err)
			return nil
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
	return ioutil.WriteFile(location, data, os.ModePerm)
}
