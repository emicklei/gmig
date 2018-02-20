package main

import (
	"encoding/json"
	"io/ioutil"
)

// ConfigFilename is for reading bucket info
const ConfigFilename = "gmig.json"

// Config holds gmig program config
type Config struct {
	// Bucket is the name of the Google Storage Bucket.
	Bucket string `json:"bucket"`

	//LastMigrationObjectName is the name of the bucket object and the local (temporary) file.
	LastMigrationObjectName string `json:"state"`

	// verbose if true then procduce more logging.
	verbose bool
}

// LoadConfig reads from gmig.json
func LoadConfig(location string) (Config, error) {
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}

// ToJSON returns the JSON representation.
func (c Config) ToJSON() string {
	data, _ := json.MarshalIndent(c, "", "\t")
	return string(data)
}
