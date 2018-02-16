package main

import (
	"log"
	"os"
	"sort"

	"github.com/emicklei/gmig"
	"github.com/urfave/cli"
)

const version = "0.1"

var stateProvider = gmig.FileStateProvider{}

func main() {
	app := cli.NewApp()
	app.Version = version
	app.EnableBashCompletion = true
	app.Name = "gmig"
	app.Usage = "GCP migrations tool"
	app.Commands = []cli.Command{
		{
			Name:   "new",
			Usage:  "gmig create \"create tester service account\"",
			Action: cmdCreateMigration,
		},
		{
			Name:        "up",
			Usage:       "gmig up [|file]",
			Description: "The up command runs the do section of all pending migrations in order, one after the other.",
			Action:      cmdMigrationsUp,
		},
		{
			Name:        "down",
			Usage:       "gmig down [|file]",
			Description: "The down command runs the undo section of the last applied migration only.",
			Action:      cmdMigrationsDown,
		},
		{
			Name:   "status",
			Usage:  "gmig status",
			Action: cmdMigrationsStatus,
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
