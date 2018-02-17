package gmig

import (
	"encoding/json"
	"io/ioutil"
)

// Config holds gmig program config
type Config struct {
	Bucket      string `json:"bucket"`
	Project     string `json:"project"`
	StateObject string `json:"state"`
}

// LoadConfig reads from gmig.json
func LoadConfig() (Config, error) {
	data, err := ioutil.ReadFile("gmig.json")
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}
