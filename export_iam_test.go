package main

import (
	"os"
	"testing"
	"time"
)

func TestExportProjectsIAMPolicy(t *testing.T) {
	// simulate effect of GS download old state
	os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm)
	defer os.Remove("state")

	// simulate now
	timeNow = func() time.Time { return time.Time{} }

	// cleanup generated migration
	generated := "010_exported_project_iam_policy.yaml"
	defer os.Remove(generated)

	// capture GC command
	cc := new(commandCapturer)
	cc.output = []byte(`{
		"bindings": [{
			"members": [
				"member"
			],
			"role": "role"
		}]
	}`)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "export", "project-iam-policy", "test/demo"}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}

	if m, err := LoadMigration(generated); err != nil {
		t.Fatal("unable to load generated migration", err)
	} else {
		if got, want := len(m.DoSection), 1; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
		if got, want := len(m.DoSection), 2; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
	}
}
