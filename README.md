# gmig - gcloud migrations

pronounced as `gimmick`.

Manage Google Cloud Platform infrastructure using migrations that describe incremental changes such as additions or deletions of resources.
This work is inspired by MyBatis migrations for SQL database setup.

Your gmig infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    /20180214t071402_create_some_account.yaml
    /my-staging-project
        gmig.json
    /my-production-project
        gmig.json

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.
A change must be have at least a `do` and an `undo` section.
The `do` section typically has a list of gcloud commands that create resources.
The `undo` section typically has a list of gcloud commands that deletes the same resources (in reverse order if relevant).
Each command can use the following environment variables: `$PROJECT`.

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

### init [project]

Prepares your setup for working with migrations by creating a `gmig.json` file in a project folder.

    gmig init my-production-project

You must change the file `my-production-project/gmig.json` to set the Bucket name.

    {
        "bucket":"mycompany-gmig-states",
        "state":"gmig-last-migration"
    }

If you decide to store state files of different projects in one Bucket then set the state object name to reflect this, eg. `myproject-gmig-state`.

### new [title]

Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "add storage view rol to cloudbuild account"

### status [project]

List all migrations with an indicator (applied,pending) whether is has been applied or not.

    gmig status my-production-project

### up [project]

Executes the `do` section of each pending migration compared to the last applied change to the infrastructure. 
Upon each completed migration, the `gmig-last-migration` object is updated.

    gmig up my-production-project

### down [project]

Executes one `undo` section of the last applied change to the infrastructure.
If completed then update the `gmig-last-migration` object.

    gmig down my-production-project

## Export existing infrastructure

Exporting migrations from existing infrastructure is useful when you start working with `gmig` but do not want to start from scratch.
Several sub commands are available to inspect a project and export migrations to reflect the current state.
After marking the current state in `gmig`, new migrations can be added that will bring your infrastructure to the next state.

### export project-iam-policy [project]

Generate a new migration by reading all the IAM policy binding from the current infrastructure of the project.

    gmig -v export project-iam-policy -s my-production-project

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