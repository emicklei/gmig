package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/emicklei/gmig"
	"github.com/urfave/cli"
)

const version = "0.1"

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
			Name:   "init",
			Usage:  "Create the initial configuration, if absent.",
			Action: cmdInit,
		},
		{
			Name:   "new",
			Usage:  "Create a new migration file from a template using a generated timestamp and a given title.",
			Action: cmdCreateMigration,
		},
		{
			Name:   "up",
			Usage:  "Runs the do section of all pending migrations in order, one after the other.",
			Action: cmdMigrationsUp,
		},
		{
			Name:   "down",
			Usage:  "Runs the undo section of the last applied migration only.",
			Action: cmdMigrationsDown,
		},
		{
			Name:   "status",
			Usage:  "List all migrations with details compared to the current state.",
			Action: cmdMigrationsStatus,
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func reportError(cfg gmig.Config, action string, err error) error {
	log.Printf("executing [%s] failed, see error below.\n", action)

	log.Println("checking gmig config ...")
	fmt.Println(cfg.ToJSON())

	log.Println("checking gcloud config list ...")
	cmd := exec.Command("gcloud", "config", "list")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))

	return err
}
