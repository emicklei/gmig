package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	// collect all filenames
	filenames := []string{}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		filenames = append(filenames, path)
		return nil
	})
	// old -> new
	sort.StringSlice(filenames).Sort()
	// fetch current stored state
	// apply all pending migrations
	workdir, err := os.Getwd()
	if err != nil {
		return err
	}
	m, err := gmig.LoadMigration(filepath.Join(workdir, filenames[0]))
	if err != nil {
		return err
	}
	m.Execute(m.Up)
	// save new state
	return nil
}
