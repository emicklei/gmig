package main

import (
	"encoding/json"
	"testing"
)

var cfg = `
{
	"project":"demo",
	"zone" : "zone",
	"region" : "region",
    "bucket":"bucket",
	"state": "state",
	"env" : {
		"FOO" : "BAR"
	}
}
`

func TestConfig(t *testing.T) {
	var c Config
	if err := json.Unmarshal([]byte(cfg), &c); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, want := c.EnvironmentVars["FOO"], "BAR"; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	if got, want := c.LastMigrationObjectName, "state"; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	if got, want := c.shellEnv()[3], "FOO=BAR"; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
}

func TestTryToLoadConfig(t *testing.T) {
	c, err := TryToLoadConfig("bogus")
	if err == nil {
		t.Error("should return error about non existing bogus folder/file")
	}
	t.Log(c, err)
}
