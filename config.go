package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// ConfigFilename is for reading bucket info
const ConfigFilename = "gmig.json"

// Config holds gmig program config
type Config struct {
	// Project is a GCP project name.
	Project string `json:"project"`

	// Region is a GCP region. Optional, use the default one if absent.
	Region string `json:"region"`

	// Region is a GCP zone. Optional, use the default one if absent.
	Zone string `json:"zone"`

	// Bucket is the name of the Google Storage Bucket.
	Bucket string `json:"bucket"`

	//LastMigrationObjectName is the name of the bucket object and the local (temporary) file.
	LastMigrationObjectName string `json:"state"`

	// EnvironmentVars hold additional environment values
	// that can be accessed by each command line in the Do & Undo section.
	// Note that PROJECT,REGION and ZONE are already available.
	EnvironmentVars map[string]string `json:"env"`

	// verbose if true then procduce more logging.
	verbose bool
}

// LoadConfig reads from gmig.json and validates it.
func LoadConfig(location string) (Config, error) {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	if err := c.Validate(); err != nil {
		return c, err
	}
	return c, nil
}

// ToJSON returns the JSON representation.
func (c Config) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "\t")
	return string(data)
}

// Validate checks required fields in the configuration.
func (c Config) Validate() error {
	if len(c.Project) == 0 {
		return errors.New("missing project in configuration")
	}
	if len(c.Bucket) == 0 {
		return errors.New("missing bucket in configuration")
	}
	if len(c.LastMigrationObjectName) == 0 {
		return errors.New("missing state name in configuration")
	}
	return nil
}

func (c Config) shellEnv() (envs []string) {
	envs = append(envs, "PROJECT="+c.Project, "REGION="+c.Project, "ZONE="+c.Zone)
	// now (override) with any custom values ; do not check values
	for k, v := range c.EnvironmentVars {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return
}
