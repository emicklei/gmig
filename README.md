# gmig - GCP migrations

pronounced as `gimmick`.

[![Build Status](https://travis-ci.org/emicklei/gmig.png)](https://travis-ci.org/emicklei/gmig)
[![Go Report Card](https://goreportcard.com/badge/github.com/emicklei/gmig)](https://goreportcard.com/report/github.com/emicklei/gmig)
[![GoDoc](https://godoc.org/github.com/emicklei/gmig?status.svg)](https://godoc.org/github.com/emicklei/gmig)

Manage Google Cloud Platform (GCP) infrastructure using migrations that describe incremental changes such as additions or deletions of resources.
This work is inspired by MyBatis migrations for SQL database setup.

[Introduction blog post](http://ernestmicklei.com/2018/03/introducing-gmig-infrastructure-as-code-for-gcp/)

Your `gmig` infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    /20180214t071402_create_some_account.yaml
    /my-staging-project
        gmig.json
    /my-production-project
        gmig.json

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.

    # add loadrunner service account

    do:
    - gcloud iam service-accounts create loadrunner --display-name "LoadRunner"

    undo:
    - gcloud iam service-accounts delete loadrunner

A change must have at least a `do` section and optionally an `undo` section.
The `do` section typically has a list of gcloud commands that create resources. Each line will be executed as a shell command so any available tool can be used.
The `undo` section typically has an ordered list of gcloud commands that deletes the same resources (in reverse order if relevant).
Each command can use the following environment variables: `$PROJECT`,`$REGION`,`$ZONE` and any additional environment variables populated from the target configuration (see `env` section in the configuration below).

Information about the last applied change to a project is stored as a Google Storage Bucket object.

## Help

    NAME:
    gmig - Google Cloud Platform infrastructure migration tool

    USAGE:
    gmig [global options] command [command options] [arguments...]

    COMMANDS:
        init     Create the initial configuration, if absent.
        new      Create a new migration file from a template using a generated timestamp and a given title.
        up       Runs the do section of all pending migrations in order, one after the other.
        down     Runs the undo section of the last applied migration only.
        status   List all migrations with details compared to the current state.
        force    state | do | undo
        export   project-iam-policy | storage-iam-policy
        help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    -q                   quiet mode, accept any prompt
    -v                   verbose logging
    --help, -h           show help
    --print-version, -V  print only the version

## Getting started

### Instalation

Currently, no pre-compiled binaries are available for download (or via a package manager) so you need to compile it using the [Go SDK](https://golang.org/dl/).

    go get github.com/emicklei/gmig

### init [path]

Prepares your setup for working with migrations by creating a `gmig.json` file in a target folder.

    gmig init my-production-project

You must change the file `my-production-project/gmig.json` to set the Bucket name.

    {
        "project": "my-production-project",
        "region": "europe-west1",
        "zone": "europe-west1-b",
        "bucket":"mycompany-gmig-states",
        "state":"gmig-last-migration",
        "env" : {
            "FOO" : "bar"
        }
    }

If you decide to store state files of different projects in one Bucket then set the state object name to reflect this, eg. `myproject-gmig-state`.
If you want to apply the same migrations to different regions/zones then choose a target folder name to reflect this, eg. `my-production-project-us-east`. Values for `region` and `zone` are required if you want to create Compute Engine resources.

### new [title]

Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "add storage view rol to cloudbuild account"

### status [path]

List all migrations with an indicator (applied,pending) whether is has been applied or not.

    gmig status my-production-project/

### up [path]

Executes the `do` section of each pending migration compared to the last applied change to the infrastructure. 
Upon each completed migration, the `gmig-last-migration` object is updated in the bucket.

    gmig up my-production-project

### down [path]

Executes one `undo` section of the last applied change to the infrastructure.
If completed then update the `gmig-last-migration` object.

    gmig down my-production-project

## Export existing infrastructure

Exporting migrations from existing infrastructure is useful when you start working with `gmig` but do not want to start from scratch.
Several sub commands are (or will become) available to inspect a project and export migrations to reflect the current state.
After marking the current state in `gmig` (using `force-state`), new migrations can be added that will bring your infrastructure to the next state.
The generated migration can ofcourse also be used to just copy commands to your own migration.

### export project-iam-policy [path]

Generate a new migration by reading all the IAM policy bindings from the current infrastructure of the project.

    gmig -v export project-iam-policy my-project/

### export storage-iam-policy [path]

Generate a new migration by reading all the IAM policy bindings, per Google Storage Bucket owned by the project.

    gmig -v export storage-iam-policy my-project/

## Working around migrations
Sometimes you need to fix things because you made a mistake or want to reorganise your work. Use the `force` and confirm your action.

### force state [path] [filename]

Explicitly set the state for the target to the last applied filename. This command can be useful if you need to working from existing infrastructure. Effectively, this filename is written to the bucket object.
Use this command with care!.

    gmig force state my-production-project 20180214t071402_create_some_account.yaml

### force do [path] [filename]

Explicitly run the commands in de `do` section of a given migration filename.
The `gmig-last-migration` object is `not` updated in the bucket.
Use this command with care!.

    gmig force do my-production-project 20180214t071402_create_some_account.yaml

### force undo [path] [filename]

Explicitly run the commands in de `undo` section of a given migration filename.
The `gmig-last-migration` object is `not` updated in the bucket.
Use this command with care!.

    gmig force undo my-production-project 20180214t071402_create_some_account.yaml

## Examples

### Create Cloud SQL Database

    # Create database

    do:
    # Standard tiers are not working for some reason using the CLI. It works using the UI
    # Note regarding the name. If already used, then cannot be used again for some time: https://github.com/hashicorp/terraform/issues/4557
    - gcloud beta sql instances create my-db --database-version=POSTGRES_9_6 --region=europe-west1 --gce-zone=europe-west1-b --availability-type=REGIONAL --cpu=1 --memory=4GB

    undo:
    - gcloud beta sql instances delete my-db

### Add Storage Viewer role

    # allow loadrunner to access GCS

    # https://cloud.google.com/iam/docs/understanding-roles#predefined_roles

    do:
    - gcloud projects add-iam-policy-binding $PROJECT --member serviceAccount:loadrunner@$PROJECT.iam.gserviceaccount.com
    --role roles/storage.objectViewer

    undo:
    - gcloud projects remove-iam-policy-binding $PROJECT --member serviceAccount:loadrunner@$PROJECT.iam.gserviceaccount.com
    --role roles/storage.objectViewer

### Add Cloud KMS CryptoKey Decrypter to cloudbuilder account

    # let cloudbuilder decrypt secrets for deployment

    # https://cloud.google.com/kms/docs/iam
    # https://cloud.google.com/kms/docs/reference/permissions-and-roles

    do:
    - gcloud kms keys add-iam-policy-binding CRYPTOKEY --location LOCATION --keyring KEYRING --member serviceAccount:00000000@cloudbuild.gserviceaccount.com --role roles/cloudkms.cryptoKeyDecrypter

    undo:
    - gcloud kms keys remove-iam-policy-binding CRYPTOKEY --location LOCATION --keyring KEYRING --member serviceAccount:00000000@cloudbuild.gserviceaccount.com --role roles/cloudkms.cryptoKeyDecrypter

&copy; 2018, ernestmicklei.com. MIT License. Contributions welcome.
