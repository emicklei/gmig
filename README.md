# gmig - GCP migrations

pronounced as `gimmick`.

[![Build Status](https://travis-ci.org/emicklei/gmig.png)](https://travis-ci.org/emicklei/gmig)
[![Go Report Card](https://goreportcard.com/badge/github.com/emicklei/gmig)](https://goreportcard.com/report/github.com/emicklei/gmig)
[![GoDoc](https://godoc.org/github.com/emicklei/gmig?status.svg)](https://godoc.org/github.com/emicklei/gmig)

Manage Google Cloud Platform (GCP) infrastructure using migrations that describe incremental changes such as additions or deletions of resources.
This work is inspired by MyBatis migrations for SQL database setup.

[Introduction blog post](http://ernestmicklei.com/2018/03/introducing-gmig-infrastructure-as-code-for-gcp/)

Your `gmig` infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    /010_create_some_account.yaml
    /015_add_permissions_to_some_account.yaml
    /my-gcp-production-project
        gmig.yaml

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.

    # add loadrunner service account

    do:
    - gcloud iam service-accounts create loadrunner --display-name "LoadRunner"

    undo:
    - gcloud iam service-accounts delete loadrunner

A change must have at least a `do` section and optionally an `undo` section.
The `do` section typically has a list of gcloud commands that create resources but any available tool can be used.
All lines will be executed at once using a single temporary shell script so you can use shell variables to simplify each section.
The `undo` section typically has an ordered list of gcloud commands that deletes the same resources (in reverse order if relevant).
Each command in each section can use the following environment variables: `$PROJECT`,`$REGION`,`$ZONE` and any additional environment variables populated from the target configuration (see `env` section in the configuration below).

## State

Information about the last applied migration to a project is stored as a Google Storage Bucket object.
Therefore, usage of this tool requires you to have create a Bucket and set the permissions (Storage Writer) accordingly. 
To view the current state of your infrastructure related to each migration, you can add the `view` section to the YAML file, such as:

    # add loadrunner service account

    do:
    - gcloud iam service-accounts create loadrunner --display-name "LoadRunner"
    
    undo:
    - gcloud iam service-accounts delete loadrunner
    
    view:
    - gcloud iam service-accounts describe loadrunner

and use the `view` subcommand.

## Conditional migration

Sometimes environments do `not` need to be setup exactly the same way.
Commands (do,undo,view) can be made conditional by adding an `if` section.
You can only use custom environment variables and configuration parameters (PROJECT,ZONE,REGION).

    if: PROJECT == "your-project-id"
    do:
    - gcloud condig list

or with combinations:

    if: (PROJECT == "your-project-id") && (ZONE == "my-zone")
    do:
    - gcloud condig list

For available operators, see [Language-Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md)

## Help

    NAME:
    gmig - Google Cloud Platform infrastructure migration tool

    USAGE:
    gmig [global options] command [command options] [arguments...]

    COMMANDS:
        init     Create the initial configuration, if absent.
        new      Create a new migration file from a template using a generated timestamp and a given title.
        up       Runs the do section of all pending migrations in order, one after the other.
                 If a migration file is specified then stop after applying that one.
        down     Runs the undo section of the last applied migration only.
        plan     Log commands of the do section of all pending migrations in order, one after the other.
        status   List all migrations with details compared to the current state.
        view     Runs the view section of all applied migrations to see the current state reported by your infrastructure.
        force    state | do | undo
        util     create-named-port | delete-named-port
        export   project-iam-policy | storage-iam-policy
        help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    -q                   quiet mode, accept any prompt
    -v                   verbose logging
    --help, -h           show help
    --print-version, -V  print only the version

## Getting started

### Installation

Pre-compiled binaries are available for download from the Releases section on this page.
If you want to create your own version, you need to compile it using the [Go SDK](https://golang.org/dl/).

    go get github.com/emicklei/gmig

### init \<path>

Prepares your setup for working with migrations by creating a `gmig.json` file in a target folder.

    gmig init my-gcp-production-project

Then your filesystem will have:

    /my-gcp-production-project/
        gmig.yaml

You must change the file `gmig.yaml` to set the Project and Bucket name.

    # gmig configuration file
    #
    # Google Cloud Platform migrations tool for infrastructure-as-code. See https://github.com/emicklei/gmig.
    
    # [project] must be the Google Cloud Project ID where the infrastructure is created.
    # Its value is available as $PROJECT in your migrations.
    #
    # Required by gmig.
    project: my-project
    
    # [region] must be a valid GCP region. See https://cloud.google.com/compute/docs/regions-zones/
    # A region is a specific geographical location where you can run your resources.
    # Its value is available as $REGION in your migrations.
    #
    # Not required by gmig but some gcloud and gsutil commands do require it.
    # region: europe-west1
    
    # [zone] must be a valid GCP zone. See https://cloud.google.com/compute/docs/regions-zones/
    # Each region has one or more zones; most regions have three or more zones.
    # Its value is available as $ZONE in your migrations.
    #
    # Not required by gmig but some gcloud and gsutil commands do require it.
    # zone: europe-west1-b
    
    # [bucket] must be a valid GCP Storage bucket.
    # A Google Storage Bucket is used to store information (object) about the last applied migration.
    # Bucket can contain multiple objects from multiple applications. Make sure the [state] is different for each app.
    #
    # Required by gmig.
    bucket: my-bucket
    
    # [state] is the name of the object that hold information about the last applied migration.
    # Required by gmig.
    state: myapp-gmig-last-migration
    
    # [env] are additional environment values that are available to each section of a migration file.
    # This can be used to create migrations that are independent of the target project.
    # By convention, use capitalized words for keys.
    # In the example, "myapp-cluster" is available as $K8S_CLUSTER in your migrations.
    #
    # Not required by gmig.
    env:
      K8S_CLUSTER: myapp-cluster

If you decide to store state files of different projects in one Bucket then set the state object name to reflect this, eg. `myproject-gmig-state`.
If you want to apply the same migrations to different regions/zones then choose a target folder name to reflect this, eg. `my-gcp-production-project-us-east`. Values for `region` and `zone` are required if you want to create Compute Engine resources. The `env` map can be used to parameterize commands in your migrations. In the example, all commands will have access to the value of `$K8S_CLUSTER`.

### new \<title>

Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "add storage view role to cloudbuild account"

Using a combination of the options `--do`, `--undo` and `--view`, you can set the commands directly for the new migration.

### status \<path> [--migrations folder]

List all migrations with an indicator (applied,pending) whether is has been applied or not.

    gmig status my-gcp-production-project/

Run this command in the directory where all migrations are stored. Use `--migrations` for a different location.

### plan \<path> [stop] [--migrations folder]

Log commands of the `do` section of all pending migrations in order, one after the other.
If `stop` is given, then stop after that migration file.

### up \<path> [stop] [--migrations folder]

Executes the `do` section of each pending migration compared to the last applied change to the infrastructure.
If `stop` is given, then stop after that migration file.
Upon each completed migration, the `gmig-last-migration` object is updated in the bucket.

    gmig up my-gcp-production-project

### down \<path> [--migrations folder]

Executes one `undo` section of the last applied change to the infrastructure.
If completed then update the `gmig-last-migration` object.

    gmig down my-gcp-production-project

### view \<path> [migration file] [--migrations folder]

Executes the `view` section of each applied migration to the infrastructure.
If `migration file` is given then run that view only.

    gmig view my-gcp-production-project

### template [-w] source-file

Processes the source-file as a Go template and write the result to stdout.
If the `-w` is given then rewrite the source with the processed content.
The following functions are available:

#### env

This function takes the first argument and does a lookup in the available OS environment values.
Example of a configuration snippet that needs the environment dependent value for $PROJECT.

    project: {{ env "PROJECT" }}

## Export existing infrastructure

Exporting migrations from existing infrastructure is useful when you start working with `gmig` but do not want to start from scratch.
Several sub commands are (or will become) available to inspect a project and export migrations to reflect the current state.
After marking the current state in `gmig` (using `force-state`), new migrations can be added that will bring your infrastructure to the next state.
The generated migration can ofcourse also be used to just copy commands to your own migration.

### export project-iam-policy \<path>

Generate a new migration by reading all the IAM policy bindings from the current infrastructure of the project.

    gmig -v export project-iam-policy my-project/

### export storage-iam-policy \<path>

Generate a new migration by reading all the IAM policy bindings, per Google Storage Bucket owned by the project.

    gmig -v export storage-iam-policy my-project/

## Working around migrations

Sometimes you need to fix things because you made a mistake or want to reorganise your work. Use the `force` and confirm your action.

### force state \<path> \<filename>

Explicitly set the state for the target to the last applied filename. This command can be useful if you need to work from existing infrastructure. Effectively, this filename is written to the bucket object.
Use this command with care!.

    gmig force state my-gcp-production-project 010_create_some_account.yaml

### force do \<path> \<filename>

Explicitly run the commands in the `do` section of a given migration filename.
The `gmig-last-migration` object is `not` updated in the bucket.
Use this command with care!.

    gmig force do my-gcp-production-project 010_create_some_account.yaml

### force undo \<path> \<filename>

Explicitly run the commands in the `undo` section of a given migration filename.
The `gmig-last-migration` object is `not` updated in the bucket.
Use this command with care!.

    gmig force undo my-gcp-production-project 010_create_some_account.yaml

## export-env \<path>
Export all available environment variable from the configuration file and also export $PROJECT, $REGION and $ZONE
Use this command with care!.

    eval $(gmig export-env my-gcp-production-project)

## GCP utilities

### util create-named-port \<instance-group> \<name:port>

The Cloud SDK has a command to [set-named-ports](https://cloud.google.com/sdk/gcloud/reference/compute/instance-groups/set-named-ports) but not a command to add or delete a single name:port mapping. To simplify the migration command for creating a name:port mapping, this gmig util command is added.
First it calls `get-named-ports` to retrieve all existing mappings. Then it will call `set-named-ports` with the new mapping unless it already exists.

### util delete-named-port \<instance-group> \<name:port>

The Cloud SDK has a command to [set-named-ports](https://cloud.google.com/sdk/gcloud/reference/compute/instance-groups/set-named-ports) but not a command to add or delete a single name:port mapping. To simplify the migration command for deleting a name:port mapping, this gmig util command is added.
First it calls `get-named-ports` to retrieve all existing mappings. Then it will call `set-named-ports` without the mapping.

## Examples

This repository has a number of [examples](https://github.com/emicklei/gmig/tree/master/examples) of migrations.

&copy; 2018, ernestmicklei.com. MIT License. Contributions welcome.
