# gmig - gcloud migrations

pronounced as `gimmick`.

Manage Google Cloud Platform infrastructure using migrations that describe incremental changes such as additions or deletions of resources. This work is inspired by MyBatis migrations for SQL database setup.

Your gmig configuration is basically a folder with change files, each with a timestamp prefix (for sort ordering) and readable name.

    \migrations
        \20180214t071402_create_some_account.yaml

Each change is a single YAML file with one or more gcloud commands that change infrastructure for a project.
A change must be have at least an `up` and a `down` section. The `up` section typically has gcloud commands that create resources.

Information about the last applied change to a project is stored in a Google Storage Bucket file.

## Example: service account
This migration uses [gcloud create service account](https://cloud.google.com/sdk/gcloud/reference/iam/service-accounts/create)

    up:
    - gcloud iam service-accounts create some-account-name --display-name "My Service Account"
    down:
    - gcloud iam service-accounts delete some-account-name

## Getting started

### gmig init
Prepares your setup for working the migrations. It checks the read/write permissions of your Bucket containing the `gmig.state` file.

### gmig new
Creates a new file for you to describe a change to the current state of infrastructure.

    gmig new "set view permissions for cloudbuild account"

### gmig status
List all migrations with an indicator whether is has been applied or not.

### gmig up
Calls the up section compared to the last applied change to the infrastructure. If completed then update the `gmig.state`. file.

### gmig down
Calls the down section of the last applied change to the infrastructure. If completed then update the `gmig.state`.

### gmig export service-accounts
Generates the YAML files by exporting from existing infrastructure of a project (creation of service accounts and setting IAM policies)

&copy; 2018, ernestmicklei.com