package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ConfigFilename is for reading bucket info
const jsonConfigFilename = "gmig.json"
const YAMLConfigFilename = "gmig.yaml"
const ymlConfigFilename = "gmig.yml"

// Config holds gmig program config
type Config struct {
	// Project is a GCP project name.
	Project string `json:"project" yaml:"project"`

	// Region is a GCP region. Optional, use the default one if absent.
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Region is a GCP zone. Optional, use the default one if absent.
	Zone string `json:"zone,omitempty" yaml:"zone,omitempty"`

	// Bucket is the name of the Google Storage Bucket.
	Bucket string `json:"bucket" yaml:"bucket"`

	//LastMigrationObjectName is the name of the bucket object and the local (temporary) file.
	LastMigrationObjectName string `json:"state" yaml:"state"`

	// EnvironmentVars hold additional environment values
	// that can be accessed by each command line in the Do & Undo section.
	// Note that PROJECT,REGION and ZONE are already available.
	EnvironmentVars map[string]string `json:"env,omitempty" yaml:"env,omitempty"`

	// verbose if true then produce more logging.
	verbose bool

	// source filename
	filename string
}

func loadAndUnmarshalConfig(location string, unmarshaller func(in []byte, out interface{}) (err error)) (*Config, error) {
	data, err := ioutil.ReadFile(location)

	if err != nil {
		return nil, err
	}

	c := &Config{
		filename: location,
	}

	if err := unmarshaller(data, &c); err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func loadYAMLConfig(location string) (*Config, error) {
	return loadAndUnmarshalConfig(location, yaml.Unmarshal)
}

func loadJSONConfig(location string) (*Config, error) {
	return loadAndUnmarshalConfig(location, json.Unmarshal)
}

// TryToLoadConfig reads configuration from path first looking for gmig.yaml,
// if not exists fallback to gmig.yml and gmig.json then validates it.
func TryToLoadConfig(pathToConfig string) (*Config, error) {
	yamlLocation := filepath.Join(pathToConfig, YAMLConfigFilename)
	ymlLocation := filepath.Join(pathToConfig, ymlConfigFilename)
	jsonLocation := filepath.Join(pathToConfig, jsonConfigFilename)

	if checkExists(yamlLocation) == nil {
		return loadYAMLConfig(yamlLocation)
	} else if checkExists(ymlLocation) == nil {
		return loadYAMLConfig(ymlLocation)
	} else if checkExists(jsonLocation) == nil {
		config, err := loadJSONConfig(jsonLocation)
		printWarning("JSON configuration (gmig.json) is deprecated, your configuration (gmig.yaml)")
		return config, err
	}

	return nil, errors.New("can not find any configuration")
}

// ToJSON returns the JSON representation.
func (c Config) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "\t")
	return string(data)
}

// ToYAML returns the YAML representation.
func (c Config) ToYAML() string {
	data, _ := yaml.Marshal(c)
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
	envs = append(envs, "PROJECT="+c.Project, "REGION="+c.Region, "ZONE="+c.Zone)
	// now (override) with any custom values ; do not check values
	for k, v := range c.EnvironmentVars {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return
}
