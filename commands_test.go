package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCmdInit(t *testing.T) {
	defer os.RemoveAll("test/new")
	if err := newApp().Run([]string{"gmig", "init", "test/new"}); err != nil {
		t.Fatal("unexpected error", err)
	}
}

func TestCmdInitMissingConfig(t *testing.T) {
	defer os.RemoveAll("test/missing")
	if err := newApp().Run([]string{"gmig", "init", "test/missing"}); err != nil {
		t.Fatal("unexpected error", err)
	}
	if _, err := os.Stat("test/missing"); err != nil {
		t.Fatal("unexpected error", err)
	}
	if _, err := os.Stat("test/missing/gmig.yaml"); err != nil {
		t.Fatal("unexpected error", err)
	}
	if err := newApp().Run([]string{"gmig", "init", "test/missing"}); err != nil {
		t.Fatal("unexpected error", err)
	}
}

func TestCmdStatusDemo(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")

	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "-v", "status", "test/demo"}); err != nil {
		t.Fatal("unexpected error", err)
	}
	if got, want := len(cc.args), 2; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	for i, each := range []string{"gcloud", "config", "set", "core/project", "demo"} {
		if got, want := cc.args[0][i], each; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
	}
	for i, each := range []string{"gsutil", "-q", "cp", "gs://bucket/state", "state"} {
		if got, want := cc.args[1][i], each; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
	}
}

func TestCmdStatusDemoWithMigrationsOption(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")

	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "-v", "status", "test/demo", "--migrations", "test"}); err != nil {
		t.Fatal("unexpected error", err)
	}
}

func TestCmdForceState(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")

	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	// do not remove state because we need to inspect it
	removeCount := 0
	osRemove = func(string) error { removeCount++; return nil }
	defer func() { osRemove = os.Remove }()

	newState := "020_two.yaml"
	if err := newApp().Run([]string{"gmig", "-q", "force", "state", "test/demo", newState}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := removeCount, 2; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	data, err := os.ReadFile("state")
	if err != nil {
		abs, _ := filepath.Abs("state")
		t.Fatal("unreadable state", abs, err)
	}
	if got, want := string(data), newState; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	for i, each := range []string{"gsutil", "-q", "-h", "Content-Type:text/plain", "cp", "state", "gs://bucket/state"} {
		if got, want := cc.args[2][i], each; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
	}
}

func TestCmdForceStateNested(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")

	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	// do not remove state because we need to inspect it
	removeCount := 0
	osRemove = func(string) error { removeCount++; return nil }
	defer func() { osRemove = os.Remove }()

	newState := "020_two.yaml"
	if err := newApp().Run([]string{"gmig", "-v", "-q", "force", "state", "--migrations", "test", "test/demo/nested", newState}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := removeCount, 2; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	data, err := os.ReadFile("state")
	if err != nil {
		abs, _ := filepath.Abs("state")
		t.Fatal("unreadable state", abs, err)
	}
	if got, want := string(data), newState; got != want {
		t.Logf("got [%v] want [%v]", got, want)
	}
	for i, each := range []string{"gsutil", "-q", "-h", "Content-Type:text/plain", "cp", "state", "gs://bucket/state"} {
		if got, want := cc.args[2][i], each; got != want {
			t.Logf("got [%v] want [%v]", got, want)
		}
	}
}

func TestCmdUp(t *testing.T) {
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	cc.err = errors.New("shell error")
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo"}); err == nil {
		wd, _ := os.Getwd()
		t.Error("expected error", err, wd)
	}
	if got, want := len(cc.args), 4; got != want { // set config, load 1, save 2, save 3, did not succeed apply error
		t.Logf("got [%v] want [%v]", got, want)
	}
}

func TestCmdUpAndStop(t *testing.T) {
	keepState()
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	cc.output = []byte("error")
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo", "020_two.yaml"}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := len(cc.args), 6; got != want { // set config, load 1, do, save 2, stop
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestCmdUpAndStopAfterLastApplied(t *testing.T) {
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("030_three.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	cc.err = errors.New("shell error")
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo", "020_two.yaml"}); err == nil {
		wd, _ := os.Getwd()
		t.Fatal("expected error", err, wd)
	}
}

func TestCmdUpAndStopAfterUnexistingFilename(t *testing.T) {
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo", "missing.yaml"}); err == nil {
		wd, _ := os.Getwd()
		t.Fatal("expected error", err, wd)
	}
	if got, want := len(cc.args), 2; got != want { // set config, load 1
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func keepState() {
	osTempDir = func() string { return "." }
	osRemove = func(s string) error { return nil } // do not remove because we run status after up
}

func TestCmdDown(t *testing.T) {
	keepState()
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "-v", "down", "test/demo"}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := len(cc.args), 3; got != want { // set config, load state 2, save state 1
		t.Logf("got [%v] want [%v]", got, want)
	}
}

func TestCmdDownWhenNoLastMigration(t *testing.T) {
	keepState()
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte(""), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "down", "test/demo"}); err != nil {
		expected := "gmig ABORTED"
		if err.Error() != expected {
			t.Errorf("got [%v] want [%v]", err, expected)
		}
	}
}

func TestCmdView(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("010_one.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "view", "test/demo"}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := len(cc.args), 3; got != want { // set config, load state, echo 3
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestCmdUpConditional(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("040_error.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo", "050_conditional.yaml"}); err != nil {
		wd, _ := os.Getwd()
		t.Fatal("unexpected error", err, wd)
	}
	if got, want := len(cc.args), 6; got != want {
		t.Log(cc.args)
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestCmdUpConditionalFail(t *testing.T) {
	osTempDir = func() string { return "." }
	// simulate effect of GS download old state
	if err := os.WriteFile("state", []byte("050_conditional.yaml"), os.ModePerm); err != nil {
		t.Fatal("unable to write state", err)
	}
	defer os.Remove("state")
	// capture GC command
	cc := new(commandCapturer)
	runCommand = cc.runCommand
	if err := newApp().Run([]string{"gmig", "up", "test/demo"}); err == nil {
		wd, _ := os.Getwd()
		t.Fatal("expected error", err, wd)
	}
	if got, want := len(cc.args), 2; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
