# gmig - gcloud migrations

pronounced as `gimmick`.

## NOTE: this is work in progress. This is not usable until the first git tag!

Manage Google Cloud Platform infrastructure using migrations that describe incremental changes such as additions or deletions of resources. This work is inspired by MyBatis migrations for SQL database setup.

Your gmig infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    \migrations
        \20180214t071402_create_some_account.yaml

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.
A change must be have at least an `do` and a `undo` section. The `do` section typically has gcloud commands that create resources.

Information about the last applied change to a project is stored in a Google Storage Bucket object.

## Example: service account
This migration uses [gcloud create service account](https://cloud.google.com/sdk/gcloud/reference/iam/service-accounts/create)

    do:
    - gcloud iam service-accounts create some-account-name --display-name "My Service Account"
    
    undo:
    - gcloud iam service-accounts delete some-account-name

## Getting started

### gmig init
Prepares your setup for working with migrations by creating a `gmig.json` file if absent.
It checks the read/write permissions of your Bucket containing the `.gmig-last-migration` object which will contain the name of the last applied migration.

    gmig init project=your-gcp-project bucket=your-bucket-name


### gmig new
Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "set view permissions for cloudbuild account"


### gmig status
List all migrations with an indicator (applied,pending) whether is has been applied or not.


### gmig up
Calls all the `do` sections of each pending migration compared to the last applied change to the infrastructure. Upon each completed migration, the `.gmig-last-migration` object is updated.


### gmig down
Calls one `undo` section of the last applied change to the infrastructure. If completed then update the `.gmig-last-migration` object.


&copy; 2018, ernestmicklei.com. MIT License. Contributions welcome.