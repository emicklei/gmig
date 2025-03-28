package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
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

func TestSetupShellScriptNotVerbose(t *testing.T) {

	want := `#!/bin/bash
# temporary gmig execution script
set -e -v`

	if got := setupShellScript(false); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestSetupShellScriptVerbose(t *testing.T) {

	want := `#!/bin/bash
# temporary gmig execution script
set -e -x`

	if got := setupShellScript(true); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestNewFilenameWithIndex(t *testing.T) {
	wd, _ := os.Getwd()
	dir, err := ioutil.TempDir("", "testing")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up
	// change and restore finally
	if err := os.Chdir(dir); err != nil {
		return
	}
	defer os.Chdir(wd)
	desc := "first migration"
	want := "010_first_migration.yaml"
	if got := NewFilenameWithIndex(desc); got != want {
		t.Errorf("NewFilenameWithIndex(%s) = %v, want %v", desc, got, want)
	}
	tmpfn := filepath.Join(dir, "20181026t183700_starts_with_timestamp.yaml")
	if err := ioutil.WriteFile(tmpfn, []byte(""), 0444); err != nil {
		log.Fatal(err)
	}
	desc = "first after timestamp"
	want = "300_first_after_timestamp.yaml"
	if got := NewFilenameWithIndex(desc); got != want {
		t.Errorf("NewFilenameWithIndex(%s) = %v, want %v", desc, got, want)
	}
	tmpfn = filepath.Join(dir, "400_starts_with_high_index.yaml")
	if err := ioutil.WriteFile(tmpfn, []byte(""), 0444); err != nil {
		log.Fatal(err)
	}
	desc = "first after high index"
	want = "405_first_after_high_index.yaml"
	if got := NewFilenameWithIndex(desc); got != want {
		t.Errorf("NewFilenameWithIndex(%s) = %v, want %v", desc, got, want)
	}
	tmpfn = filepath.Join(dir, "unexpected_yaml_in_directory.yaml")
	if err := ioutil.WriteFile(tmpfn, []byte(""), 0444); err != nil {
		log.Fatal(err)
	}
	desc = "potentially unexpected naming"
	want = "010_potentially_unexpected_naming.yaml"
	if got := NewFilenameWithIndex(desc); got != want {
		t.Errorf("NewFilenameWithIndex(%s) = %v, want %v", desc, got, want)
	}
}

func TestEvaluateCondition(t *testing.T) {
	envs := []string{"ZONE=A", "PROJECT=B"}
	ok, err := evaluateCondition(`PROJECT == "B"`, envs)
	if err != nil {
		log.Fatal(err)
	}
	if got, want := ok, true; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestMigrationContextWithMigrationsOverride(t *testing.T) {
	t.Skip()
	c := new(cli.Context)
	c.Set("migrations", "/tmp/here")
	ctx, err := getMigrationContext(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ctx)
}
