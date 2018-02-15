package main

import (
	"io/ioutil"
	"os"

	"github.com/emicklei/gmig"
	"github.com/urfave/cli"
)

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
	upToIncluding := c.Args.First()
	// load old state
	first := gmig.LoadMigrationsBetweenAnd("", upToIncluding)
	first.Execute(first.Up)
	// save new state
	return nil
}
