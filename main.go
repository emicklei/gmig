package main

import (
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

const version = "latest"

func main() {
	if err := newApp().Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func newApp() *cli.App {
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
		cli.BoolFlag{
			Name:  "q",
			Usage: "quiet mode, accept any prompt",
		},
	}
	migrationsFlag := cli.StringFlag{
		Name: "migrations",
		Usage: `folder containing the migrations to apply on the target project.
	If not specified then set it to the parent folder of the configuration file.`,
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Create the initial configuration, if absent.",
			Action: func(c *cli.Context) error {
				defer started(c, "init")()
				return cmdInit(c)
			},
			ArgsUsage: `<path>
				path - name of the folder that contains the configuration of the target project. The folder name may end with a path separator and can be relative or absolute.`,
		},
		{
			Name:  "new",
			Usage: "Create a new migration file from a template using an index prefix and a given title.",
			Action: func(c *cli.Context) error {
				defer started(c, "create migration")()
				return cmdCreateMigration(c)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "do",
					Usage: "commands to run in the 'do' section of this migration. Multiple commands need to be separated by a newline.",
				},
				cli.StringFlag{
					Name:  "undo",
					Usage: "commands to run in the 'undo' section of this migration. Multiple commands need to be separated by a newline.",
				},
				cli.StringFlag{
					Name:  "view",
					Usage: "commands to run in the 'view' section of this migration. Multiple commands need to be separated by a newline.",
				},
			},
			ArgsUsage: `<title>
				title - what the effect of this migration is on infrastructure.`,
		},
		{
			Name:  "plan",
			Usage: "Log commands of the do section of all pending migrations in order, one after the other. If a migration file is specified then stop after applying that one.",
			Action: func(c *cli.Context) error {
				defer started(c, "plan = log commands of pending migrations")()
				return cmdMigrationsPlan(c)
			},
			Flags: []cli.Flag{migrationsFlag},
			ArgsUsage: `<path> [stop] 
				path - name of the folder that contains the configuration of the target project.
				stop - (optional) the name of the migration file after which applying migrations will stop.`,
		},
		{
			Name:  "up",
			Usage: "Runs the do section of all pending migrations in order, one after the other. If a migration file is specified then stop after applying that one.",
			Action: func(c *cli.Context) error {
				defer started(c, "up = apply pending migrations")()
				return cmdMigrationsUp(c)
			},
			Flags: []cli.Flag{migrationsFlag},
			ArgsUsage: `<path> [stop] 
				path - name of the folder that contains the configuration of the target project.
				stop - (optional) the name of the migration file after which applying migrations will stop.`,
		},
		{
			Name:  "down",
			Usage: "Runs the undo section of only the last applied migration.",
			Action: func(c *cli.Context) error {
				defer started(c, "down = undo last applied migration")()
				return cmdMigrationsDown(c)
			},
			Flags: []cli.Flag{migrationsFlag},
			ArgsUsage: `<path>
				path - name of the folder that contains the configuration of the target project.`,
		},
		{
			Name:  "status",
			Usage: "List all migrations with details compared to the current state.",
			Action: func(c *cli.Context) error {
				defer started(c, "show status of migrations")()
				return cmdMigrationsStatus(c)
			},
			Flags: []cli.Flag{migrationsFlag},
			ArgsUsage: `<path>
				path - name of the folder that contains the configuration of the target project.`,
		},
		{
			Name:  "view",
			Usage: "Show infrastructure information for the current state.",
			Action: func(c *cli.Context) error {
				defer started(c, "show status of infrastructure")()
				return cmdView(c)
			},
			Flags: []cli.Flag{migrationsFlag},
			ArgsUsage: `<path>
				path - name of the folder that contains the configuration of the target project.`,
		},
		{
			Name:  "util",
			Usage: "Handle named ports {create-named-port|delete-named-port}",
			Subcommands: []cli.Command{
				{
					Name:  "create-named-port",
					Usage: "add a new name->port mapping to a compute instance group. ignore if exists.",
					Action: func(c *cli.Context) error {
						defer started(c, "create-named-port")()
						return cmdCreateNamedPort(c)
					},
					ArgsUsage: `<instance-group> <name:port>
					instance-group - identifier of the compute instance group
					name:port      - mapping of a name to a port, e.g  http-port:80`,
				},
				{
					Name:  "delete-named-port",
					Usage: "delete a name->port mapping from a compute instance group. ignore if not exists.",
					Action: func(c *cli.Context) error {
						defer started(c, "delete-named-port")()
						return cmdDeleteNamedPort(c)
					},
					ArgsUsage: `<instance-group> <name:port>
					instance-group - identifier of the compute instance group
					name:port      - mapping of a name to a port, e.g  http-port:80`,
				},
			},
		},
		{
			Name:  "force",
			Usage: "Force an action {state|do|undo}",
			Flags: []cli.Flag{migrationsFlag},
			Subcommands: []cli.Command{
				{
					Name:  "state",
					Usage: "Explicitly set the state to a specified migration filename.",
					Action: func(c *cli.Context) error {
						defer started(c, "force last applied migration (state)")()
						return cmdMigrationsSetState(c)
					},
					ArgsUsage: `<path>
					path - name of the folder that contains the configuration of the target project.`,
				},
				{
					Name:  "do",
					Usage: "Force run the DO section of a migration. State will not be updated.",
					Action: func(c *cli.Context) error {
						defer started(c, "execute DO section")()
						return cmdRundoOnly(c)
					},
					ArgsUsage: `<path> <filename>
						path - name of the folder that contains the configuration of the target project.
						filename - name of the migration that contains a do: section.`,
				},
				{
					Name:  "undo",
					Usage: "Force run the UNDO section of a migration. State will not be updated.",
					Action: func(c *cli.Context) error {
						defer started(c, "execute UNDO section")()
						return cmdRunUndoOnly(c)
					},
					ArgsUsage: `<path> <filename>
					path - name of the folder that contains the configuration of the target project.
					filename - name of the migration that contains a undo: section.`,
				},
			},
		},
		{
			Name:  "export",
			Usage: "Export existing infrastructure {project-iam-policy|storage-iam-policy}",
			Subcommands: []cli.Command{
				{
					Name:  "project-iam-policy",
					Usage: "Print a migration that describes the current IAM policy binding on project level.",
					Action: func(c *cli.Context) error {
						defer started(c, "export project IAM policy")()
						return cmdExportProjectIAMPolicy(c)
					},
					ArgsUsage: `<path>
					path - name of the folder that contains the configuration of the target project.`,
				},
				{
					Name:  "storage-iam-policy",
					Usage: "Print a migration that describes the current IAM policy bindings for Google Storage.",
					Action: func(c *cli.Context) error {
						defer started(c, "export storage IAM policy")()
						return cmdExportStorageIAMPolicy(c)
					},
					ArgsUsage: `<path>
					path - name of the folder that contains the configuration of the target project.`,
				},
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	return app
}
