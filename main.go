package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/urfave/cli"
)

const version = "0.5"

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
			Name:      "init",
			Usage:     "Create the initial configuration, if absent.",
			Action:    cmdInit,
			ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
		},
		{
			Name:      "new",
			Usage:     "Create a new migration file from a template using a generated timestamp and a given title.",
			Action:    cmdCreateMigration,
			ArgsUsage: "[title] what the migration achieves",
		},
		{
			Name:      "up",
			Usage:     "Runs the do section of all pending migrations in order, one after the other.",
			Action:    cmdMigrationsUp,
			ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
		},
		{
			Name:      "down",
			Usage:     "Runs the undo section of the last applied migration only.",
			Action:    cmdMigrationsDown,
			ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
		},
		{
			Name:      "status",
			Usage:     "List all migrations with details compared to the current state.",
			Action:    cmdMigrationsStatus,
			ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
		},
		{
			Name: "export",
			Subcommands: []cli.Command{
				{
					Name:      "project-iam-policy",
					Usage:     "Print a migration that describes the current IAM policy binding on project level.",
					Action:    cmdExportProjectIAMPolicy,
					ArgsUsage: "[project] name of the folder that contains the configuration of the target project",
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func reportError(cfg Config, action string, err error) error {
	log.Printf("executing [%s] failed, see error above and or below.\n", action)

	log.Println("checking gmig config ...")
	fmt.Println(cfg.ToJSON())

	log.Println("checking gcloud config list ...")
	cmd := exec.Command("gcloud", "config", "list")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))

	return err
}
