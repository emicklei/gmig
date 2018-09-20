package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"text/template"

	"github.com/go-yaml/yaml"
)

// Migration holds shell commands for applying or reverting a change.
type Migration struct {
	Filename    string   `yaml:"-"`
	Description string   `yaml:"-"`
	DoSection   []string `yaml:"do"`
	UndoSection []string `yaml:"undo"`
	ViewSection []string `yaml:"view"`
}

// for testing
var timeNow = time.Now

// NewFilename generates a filename for storing a new migration.
func NewFilename(desc string) string {
	now := timeNow()
	sanitized := strings.Replace(strings.ToLower(desc), " ", "_", -1)
	return fmt.Sprintf("%d%02d%02dt%02d%02d%02d_%s.yaml", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), sanitized)
}

// LoadMigration reads and parses a migration from a named file.
func LoadMigration(absFilename string) (m Migration, err error) {
	data, err := ioutil.ReadFile(absFilename)
	if err != nil {
		wd, _ := os.Getwd()
		return m, fmt.Errorf("in %s, %s reading failed: %v", wd, absFilename, err)
	}
	m.Filename = filepath.Base(absFilename)
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		err = fmt.Errorf("%s parsing failed: %v", absFilename, err)
	}
	return
}

// ToYAML returns the contents of a YAML encoded fixture.
func (m Migration) ToYAML() ([]byte, error) {
	out := new(bytes.Buffer)
	err := migrationTemplate.Execute(out, m)
	return out.Bytes(), err
}

// ExecuteAll the commands for this migration.
// We create a temporary executable file with all commands.
// This allows for using shell variables in multiple commands.
func ExecuteAll(commands []string, envs []string, verbose bool) error {
	if len(commands) == 0 {
		return nil
	}
	tempScript := path.Join(os.TempDir(), "gmig.sh")
	content := new(bytes.Buffer)
	fmt.Fprintln(content, setupShellScript(verbose))

	for _, each := range commands {
		fmt.Fprintln(content, each)
	}
	if err := ioutil.WriteFile(tempScript, content.Bytes(), os.ModePerm); err != nil {
		return fmt.Errorf("failed to write temporary migration section:%v", err)
	}
	defer func() {
		if err := os.Remove(tempScript); err != nil {
			log.Printf("warning: failed to remove temporary migration execution script:%s\n", tempScript)
		}
	}()
	cmd := exec.Command("sh", "-c", tempScript)
	cmd.Env = append(os.Environ(), envs...) // extend, not replace
	if out, err := runCommand(cmd); err != nil {
		return fmt.Errorf("failed to run migration section:\n%s\nerror:%v", string(out), err)
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func setupShellScript(verbose bool) string {
	flag := "-v"
	if verbose {
		flag = "-x"
	}
	return fmt.Sprintf(`#!/bin/bash
set -e %s`, flag)
}

// LoadMigrationsBetweenAnd returns a list of pending Migration <firstFilename..lastFilename]
func LoadMigrationsBetweenAnd(migrationsPath, firstFilename, lastFilename string) (list []Migration, err error) {
	// collect all filenames
	filenames := []string{}
	// firstFilename and lastFilename are relative to workdir.
	here, _ := os.Getwd()
	// change and restore finally
	if err = os.Chdir(migrationsPath); err != nil {
		return
	}
	defer os.Chdir(here)
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !isYamlFile(path) {
			return nil
		}
		filenames = append(filenames, path)
		return nil
	})
	// old -> new
	sort.StringSlice(filenames).Sort()
	// load only pending migrations
	for _, each := range filenames {
		// do not include firstFilename
		if each <= firstFilename {
			continue
		}
		var m Migration
		m, err = LoadMigration(filepath.Join(migrationsPath, each))
		if err != nil {
			return
		}
		list = append(list, m)
		// include lastFilename
		if len(lastFilename) == 0 {
			continue
		}
		if each == lastFilename {
			return
		}
	}
	return
}

var migrationTemplate = template.Must(template.New("gen").Parse(`
# {{.Description}}
#
# file: {{.Filename}}

do:{{range .DoSection}}
- {{.}}{{end}}

undo:{{range .UndoSection}}
- {{.}}{{end}}

view:{{range .ViewSection}}
- {{.}}{{end}}
`))
