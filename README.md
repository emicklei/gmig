# gmig - gcloud migrations

pronounced as `gimmick`.

Manage Google Cloud Platform infrastructure using migrations that describe incremental changes such as additions or deletions of resources. 
This work is inspired by MyBatis migrations for SQL database setup.

Your gmig infrastructure is basically a folder with incremental change files, each with a timestamp prefix (for sort ordering) and readable name.

    \migrations
        \20180214t071402_create_some_account.yaml
        \gmig.json

Each change is a single YAML file with one or more shell commands that change infrastructure for a project.
A change must be have at least a `do` and an `undo` section. 
The `do` section typically has gcloud commands that create resources.
The `undo` section typically has gcloud commands that deletes the same resources (in reverse order if applicable).

Information about the last applied change to a project is stored as a Google Storage Bucket object.



## Example: service account
This migration uses [gcloud create service account](https://cloud.google.com/sdk/gcloud/reference/iam/service-accounts/create)

    do:
    - gcloud iam service-accounts create some-account-name --display-name "My Service Account"
    
    undo:
    - gcloud iam service-accounts delete some-account-name




## Getting started

### gmig init
Prepares your setup for working with migrations by creating a `gmig.json` file if absent.

    gmig init project=your-gcp-project bucket=your-bucket-name

You must change this file to set the Bucket name. 
If you decide to store state files of different projects in one bucket then set the state object name to reflect this.


### gmig new
Creates a new migration for you to describe a change to the current state of infrastructure.

    gmig new "add storage view rol to cloudbuild account"


### gmig status
List all migrations with an indicator (applied,pending) whether is has been applied or not.


### gmig up
Executes the `do` section of each pending migration compared to the last applied change to the infrastructure. 
Upon each completed migration, the `.gmig-last-migration` object is updated.


### gmig down
Executes one `undo` section of the last applied change to the infrastructure. 
If completed then update the `.gmig-last-migration` object.


## Example: Add service account

    # add loadrunner service account

    do:
    - gcloud iam service-accounts create loadrunner --display-name "LoadRunner"

    undo:
    - gcloud iam service-accounts delete loadrunner

## Example: Add Storage Viewer role


    # allow loadrunner to access GCS

    # https://cloud.google.com/iam/docs/granting-roles-to-service-accounts

    do:
    - gcloud projects add-iam-policy-binding gmig-demo --member serviceAccount:loadrunner@gmig-demo.iam.gserviceaccount.com
    --role roles/storage.objectViewer

    undo:
    - gcloud projects remove-iam-policy-binding gmig-demo --member serviceAccount:loadrunner@gmig-demo.iam.gserviceaccount.com
    --role roles/storage.objectViewer



&copy; 2018, ernestmicklei.com. MIT License. Contributions welcome.