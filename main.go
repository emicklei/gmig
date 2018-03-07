package main

import (
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

const version = "0.15"

func main() {
	app := cli.NewApp()
	app.Version = version
	app.EnableBashCompletion = true
	app.Name = "gmig"
	app.Usage = "Google Cloud Platform infrastructure migration tool"

	// override -v
	cli.VersionFlag = cli.BoolFlag{
		Name:  "print-version, V",
		Usage: "print only the version",
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "v",
			Usage: "verbose logging",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Create the initial configuration, if absent.",
			Action: func(c *cli.Context) error {
				defer started(c, "init")()
				return cmdInit(c)
			},
			ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
		},
		{
			Name:  "new",
			Usage: "Create a new migration file from a template using a generated timestamp and a given title.",
			Action: func(c *cli.Context) error {
				defer started(c, "create migration")()
				return cmdCreateMigration(c)
			},
			ArgsUsage: "[title] what the migration achieves",
		},
		{
			Name:  "up",
			Usage: "Runs the do section of all pending migrations in order, one after the other.",
			Action: func(c *cli.Context) error {
				defer started(c, "up = apply pending migrations")()
				return cmdMigrationsUp(c)
			},
			ArgsUsage: "[target] name of the folder that contains the configuration of the target project",
		},
		{
			Name:  "down",
			Usage: "Runs the undo section of the last applied migration only.",
			Action: func(c *cli.Context) error {
				defer started(c, "down = undo last applied migration")()
				return cmdMigrationsDown(c)
			},
			ArgsUsage: "[target] name of the folder that contains the configuration of the target project",
		},
		{
			Name:  "status",
			Usage: "List all migrations with details compared to the current state.",
			Action: func(c *cli.Context) error {
				defer started(c, "show status of migrations")()
				return cmdMigrationsStatus(c)
			},
			ArgsUsage: "[target] name of the folder that contains the configuration of the target project",
		},
		{
			Name:  "force-state",
			Usage: "Explicitly set the current state; filename of the last applied migration.",
			Action: func(c *cli.Context) error {
				defer started(c, "force last applied migration (state)")()
				return cmdMigrationsSetState(c)
			},
			ArgsUsage: "[target] [filename] name of the folder that contains the configuration of the target project",
		},
		{
			Name: "export",
			Subcommands: []cli.Command{
				{
					Name:  "project-iam-policy",
					Usage: "Print a migration that describes the current IAM policy binding on project level.",
					Action: func(c *cli.Context) error {
						defer started(c, "export project IAM policy")()
						return cmdExportProjectIAMPolicy(c)
					},
					ArgsUsage: "[target] name of the folder that contains the configuration of the target project",
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
