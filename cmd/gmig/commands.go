package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/emicklei/gmig"
	"github.com/urfave/cli"
)

// space is right after timestamp
const logseparator = "~-------------- ---------------------~"

func cmdCreateMigration(c *cli.Context) error {
	desc := c.Args().First()
	filename := gmig.NewFilename(desc)
	m := gmig.Migration{
		Description: desc,
		Up:          []string{"gcloud config list"},
		Down:        []string{"gcloud config list"},
	}
	yaml, err := m.ToYAML()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(yaml), os.ModePerm)
}

func cmdMigrationsUp(c *cli.Context) error {
	lastApplied, _ := stateProvider.LoadState()
	all, err := gmig.LoadMigrationsBetweenAnd(lastApplied, c.Args().First())
	if err != nil {
		return err
	}
	for _, each := range all {
		log.Println(logseparator)
		log.Println(each.Filename)
		log.Println(logseparator)
		if err := gmig.ExecuteAll(each.Up); err != nil {
			return err
		}
		lastApplied = each.Filename
		// save after each succesful migration
		if err := stateProvider.SaveState(lastApplied); err != nil {
			return err
		}
	}
	return nil
}

func cmdMigrationsDown(c *cli.Context) error {
	lastApplied, _ := stateProvider.LoadState()
	all, err := gmig.LoadMigrationsBetweenAnd("", lastApplied)
	if err != nil {
		return err
	}
	lastMigration := all[len(all)-1]
	log.Println(logseparator)
	log.Println(lastApplied)
	log.Println(logseparator)
	if err := gmig.ExecuteAll(lastMigration.Down); err != nil {
		return err
	}
	// save after succesful migration
	previousFilename := ""
	if len(all) > 1 {
		previousFilename = all[len(all)-2].Filename
	}
	if err := stateProvider.SaveState(previousFilename); err != nil {
		return err
	}
	return nil
}

func cmdMigrationsStatus(c *cli.Context) error {
	lastApplied, _ := stateProvider.LoadState()
	all, err := gmig.LoadMigrationsBetweenAnd("", "")
	if err != nil {
		return err
	}
	log.Println(logseparator)
	var last string
	for _, each := range all {
		status := "---applied---"
		if each.Filename > lastApplied {
			status = "...pending..."
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
