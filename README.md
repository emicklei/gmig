# gmig - GCP migrations

pronounced as `gimmick`.

[![Build Status](https://travis-ci.org/emicklei/gmig.png)](https://travis-ci.org/emicklei/gmig)
[![Go Report Card](https://goreportcard.com/badge/github.com/emicklei/gmig)](https://goreportcard.com/report/github.com/emicklei/gmig)
[![GoDoc](https://godoc.org/github.com/emicklei/gmig?status.svg)](https://godoc.org/github.com/emicklei/gmig)

Manage Google Cloud Platform (GCP) infrastructure using migrations that describe incremental changes such as additions or deletions of resources.
This work is inspired by MyBatis migrations for SQL database setup.

Your gmig infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    /20180214t071402_create_some_account.yaml
    /my-staging-project
        gmig.json
    /my-production-project
        gmig.json

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.
A change must have at least a `do` section and optionally an `undo` section.
The `do` section typically has a list of gcloud commands that create resources. Each line will be executed as a shell command so any available tool can be used.
The `undo` section typically has a list of gcloud commands that deletes the same resources (in reverse order if relevant).
Each command can use the following environment variables: `$PROJECT`,`$REGION`,`$ZONE` and any additional environment variables populated from the target configuration (see `env`).

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
        export
        help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    -v                   verbose logging
    --help, -h           show help
    --print-version, -V  print only the version

## Getting started

### init [target]

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
            "ANSWER" : "42"
        }
    }

If you decide to store state files of different projects in one Bucket then set the state object name to reflect this, eg. `myproject-gmig-state`.
If you want to apply the same migrations to different regions/zones then choose a target folder name to reflect this, eg. `my-production-project-us-east`. Values for `region` and `zone` are required if you want to create Compute Engine resources.

### new [title]

Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "add storage view rol to cloudbuild account"

### status [target]

List all migrations with an indicator (applied,pending) whether is has been applied or not.

    gmig status my-production-project

### up [target]

Executes the `do` section of each pending migration compared to the last applied change to the infrastructure. 
Upon each completed migration, the `gmig-last-migration` object is updated in the bucket.

    gmig up my-production-project

### down [target]

Executes one `undo` section of the last applied change to the infrastructure.
If completed then update the `gmig-last-migration` object.

    gmig down my-production-project

### force-state [target] [filename]

Explicitly set the state for the target to the last applied filename. This command can be useful if you need to working from existing infrastructure. Effectively, this filename is written to the bucket object.

    gmig force-state my-production-project 20180214t071402_create_some_account.yaml

## Export existing infrastructure

Exporting migrations from existing infrastructure is useful when you start working with `gmig` but do not want to start from scratch.
Several sub commands are (or will become) available to inspect a project and export migrations to reflect the current state.
After marking the current state in `gmig`, new migrations can be added that will bring your infrastructure to the next state.

### export project-iam-policy [target]

Generate a new migration by reading all the IAM policy binding from the current infrastructure of the project.

    gmig -v export project-iam-policy my-production-project

Option `v` means verbose logging.

## Example: Add service account

    # add loadrunner service account

    do:
    - gcloud iam service-accounts create loadrunner --display-name "LoadRunner"

    undo:
    - gcloud iam service-accounts delete loadrunner


## Example: Add Storage Viewer role

    # allow loadrunner to access GCS

    # https://cloud.google.com/iam/docs/understanding-roles#predefined_roles

    do:
    - gcloud projects add-iam-policy-binding $PROJECT --member serviceAccount:loadrunner@$PROJECT.iam.gserviceaccount.com
    --role roles/storage.objectViewer

    undo:
    - gcloud projects remove-iam-policy-binding $PROJECT --member serviceAccount:loadrunner@$PROJECT.iam.gserviceaccount.com
    --role roles/storage.objectViewer


## Example: Add Cloud KMS CryptoKey Decrypter to cloudbuilder account

    # let cloudbuilder decrypt secrets for deployment

    # https://cloud.google.com/kms/docs/iam
    # https://cloud.google.com/kms/docs/reference/permissions-and-roles

    do:
    - gcloud kms keys add-iam-policy-binding CRYPTOKEY --location LOCATION --keyring KEYRING --member serviceAccount:00000000@cloudbuild.gserviceaccount.com --role roles/cloudkms.cryptoKeyDecrypter

    undo:
    - gcloud kms keys remove-iam-policy-binding CRYPTOKEY --location LOCATION --keyring KEYRING --member serviceAccount:00000000@cloudbuild.gserviceaccount.com --role roles/cloudkms.cryptoKeyDecrypter


&copy; 2018, ernestmicklei.com. MIT License. Contributions welcome.