package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"text/template"

	"github.com/expr-lang/expr"
	"gopkg.in/yaml.v2"
)

// Migration holds shell commands for applying or reverting a change.
type Migration struct {
	Filename     string   `yaml:"-"`
	Description  string   `yaml:"-"`
	IfExpression string   `yaml:"if"`
	DoSection    []string `yaml:"do"`
	UndoSection  []string `yaml:"undo"`
	ViewSection  []string `yaml:"view"`
}

// evaluateCondition evaluates the expression to a bool ; report error otherwise.
func evaluateCondition(ifExpression string, envs []string) (bool, error) {
	if len(ifExpression) == 0 {
		return true, nil
	}
	envMap := map[string]string{}
	for _, each := range envs {
		kv := strings.Split(each, "=")
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		envMap[k] = v
	}
	program, err := expr.Compile(ifExpression, expr.Env(envMap))
	if err != nil {
		return false, err
	}
	output, err := expr.Run(program, envMap)
	if err != nil {
		return false, err
	}
	if b, ok := output.(bool); ok {
		return b, nil
	}
	return false, errors.New("expression does not evaluate to a boolean")
}

// for testing
var timeNow = time.Now

// NewFilenameWithIndex generates a filename using an index for storing
// a new migration.
func NewFilenameWithIndex(desc string) string {
	all, err := LoadMigrationsBetweenAnd(".", "", "")
	if err != nil {
		printError(err.Error())
		return ""
	}
	sanitized := strings.Replace(strings.ToLower(desc), " ", "_", -1)
	if len(all) == 0 {
		return fmt.Sprintf("010_%s.yaml", sanitized)
	}
	lastFilename := all[len(all)-1].Filename
	hasTimestamp := regexpTimestamp.MatchString(lastFilename)
	hasIndex := regexpIndex.MatchString(lastFilename)
	if hasIndex {
		i, err := strconv.Atoi(lastFilename[:3])
		if err != nil {
			fmt.Printf("%T, %v", i, i)
		}
		return fmt.Sprintf("%03d_%s.yaml", i+5, sanitized)
	}
	if hasTimestamp {
		return fmt.Sprintf("300_%s.yaml", sanitized)
	}
	return fmt.Sprintf("010_%s.yaml", sanitized)
}

// LoadMigration reads and parses a migration from a named file.
func LoadMigration(absFilename string) (m Migration, err error) {
	data, err := os.ReadFile(absFilename)
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

// ExecuteAll the commands for this migration unless the condition evaluates to false
// We create a temporary executable file with all commands.
// This allows for using shell variables in multiple commands.
func ExecuteAll(ifExpression string, commands []string, envs []string, verbose bool) error {
	// check condition
	pass, err := evaluateCondition(ifExpression, envs)
	if err != nil {
		log.Printf("unable to evaluate condition [%s] because:%v\n", ifExpression, err)
		return errAbort
	}
	if !pass {
		log.Printf(".. skipping ... (%d) commands because %s is false.\n", len(commands), ifExpression)
		return nil
	}
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
	if verbose {
		log.Println("--- BEGIN gmig.sh:\n", content.String(), "--- END gmig.sh")
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

// LogAll logs expanded commands using the environment variables of both the config and the OS.
func LogAll(ifExpression string, commands []string, envs []string, verbose bool) error {
	// check condition
	pass, err := evaluateCondition(ifExpression, envs)
	if err != nil {
		log.Printf("unable to evaluate condition [%s] because:%v\n", ifExpression, err)
		return errAbort
	}
	if !pass {
		log.Printf(".. skipping ... (%d) commands because %s is false.\n", len(commands), ifExpression)
		return nil
	}
	if len(commands) == 0 {
		return nil
	}
	allEnv := append(os.Environ(), envs...)
	envMap := map[string]string{}
	for _, each := range allEnv {
		kv := strings.Split(each, "=")
		envMap[kv[0]] = kv[1]
	}
	for _, each := range commands {
		log.Println(expandVarsIn(envMap, each))
	}
	return nil
}

// expandVarsIn returns a command with all occurrences of environment variables replaced by known values.
func expandVarsIn(envs map[string]string, command string) string {
	// assume no recurse expand
	expanded := command
	for k, v := range envs {
		// if the value itself is a known variable then skip it
		if strings.HasPrefix(v, "$") {
			if _, ok := envs[v]; ok {
				log.Printf("Warning, skipping non-expandable environment var %s=%v\n", k, v)
				continue
			}
		}
		varName := "$" + k
		expanded = strings.Replace(expanded, varName, v, -1)
	}
	return expanded
}

func setupShellScript(verbose bool) string {
	flag := "-v"
	if verbose {
		flag = "-x"
	}
	return fmt.Sprintf(`#!/bin/bash
# temporary gmig execution script
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
	files, err := os.ReadDir(".")
	if err != nil {
		log.Println("unable to read migrations from folder", err)
		return
	}
	for _, each := range files {
		if each.IsDir() || !isYamlFile(each.Name()) {
			continue
		}
		filenames = append(filenames, each.Name())
	}
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
