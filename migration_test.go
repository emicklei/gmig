package main

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/go-yaml/yaml"
)

var one = `
do:
- going up
# comment for down
undo:
- going down
`

func TestParseMigration(t *testing.T) {
	var m Migration
	if err := yaml.Unmarshal([]byte(one), &m); err != nil {
		t.Error(err)
	}
	if got, want := m.DoSection[0], "going up"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.UndoSection[0], "going down"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

// gcloud config list --format json
func readConfig() Config {
	type gcloudconfig struct {
		Compute struct {
			Region, Zone string
		}
		Core struct {
			Project string
		}
	}
	var gc gcloudconfig
	cmd := exec.Command("gcloud", "config", "list", "--format", "json")
	out, _ := runCommand(cmd)
	json.Unmarshal(out, &gc)
	return Config{
		Project: gc.Core.Project,
		Region:  gc.Compute.Region,
		Zone:    gc.Compute.Zone,
	}
}

func TestSetupShellScriptNotVerbose(t *testing.T) {

	want := `#!/bin/bash
set -e -v`

	if got := setupShellScript(false); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestSetupShellScriptVerbose(t *testing.T) {

	want := `#!/bin/bash
set -e -x`

	if got := setupShellScript(true); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
