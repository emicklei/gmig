package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

var initialYAMLConfig = `
# gmig configuration file
#
# Google Cloud Platform migrations tool for infrastructure-as-code. See https://github.com/emicklei/gmig .

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

# [bucket] must be a valid GPC bucket.
# A Google Storage Bucket is used to store information (object) about the last applied migration.
# Bucket can contain multiple objects from multiple applications. Make sure the [state] is different for each app.
#
# Required by gmig.
bucket: my-bucket

# [state] is the name of the object that hold information about the last applied migration.
#
# Required by gmig.
state: myapp-gmig-last-migration

# [env] are additional environment values that are available to each section of a migration file.
# This can be used to create migrations that are independent of the target project.
# By convention, use capitalized words for keys.
# In the example, "myapp-cluster" is available as $K8S_CLUSTER in your migrations.
#
# Not required by gmig.
#env:
#  K8S_CLUSTER: myapp-cluster
`

func cmdInit(c *cli.Context) error {
	target := c.Args().First()
	if len(target) == 0 {
		printError("missing target name in command line")
		return errAbort
	}
	if err := os.MkdirAll(target, os.ModePerm|os.ModeDir); err != nil {
		printError(err.Error())
		return errAbort
	}
	config, err := TryToLoadConfig(target)
	if config != nil && err == nil {
		log.Println("config file [", config.filename, "] already present.")
		// TODO move to Config
		log.Println("config [ bucket=", config.Bucket, ",state=", config.LastMigrationObjectName, ",verbose=", config.verbose, "]")
		return nil
	} else if config != nil && err != nil {
		printError(err.Error())
		return errAbort
	}
	location := filepath.Join(target, YAMLConfigFilename)
	err = ioutil.WriteFile(location, []byte(initialYAMLConfig), os.ModePerm)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	return nil
}
