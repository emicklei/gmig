package gmig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
)

// Migration holds shell commands for applying or reverting a change.
type Migration struct {
	Filename    string   `yaml:"-"`
	Description string   `yaml:"-"`
	Up          []string `yaml:"up"`
	Down        []string `yaml:"down"`
}

// NewFilename generates a filename for storing a new migration.
func NewFilename(desc string) string {
	now := time.Now()
	sanitized := strings.Replace(strings.ToLower(desc), " ", "_", -1)
	return fmt.Sprintf("%d%02d%02dt%02d%02d%02d_%s.yaml", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), sanitized)
}

// LoadMigration reads and parses a migration from a named file.
func LoadMigration(filename string) (m Migration, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return m, err
	}
	err = yaml.Unmarshal(data, &m)
	return
}

// ToYAML returns the contents of a YAML encoded fixture.
func (m Migration) ToYAML() ([]byte, error) {
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "# %s\n\n", m.Description)
	data, err := yaml.Marshal(m)
	if err != nil {
		return data, err
	}
	out.Write(data)
	return out.Bytes(), nil
}

// Execute the commands for this migration.
func (m Migration) Execute(commands []string) error {
	if len(commands) == 0 {
		return nil
	}
	for i, each := range commands {
		log.Println("sh -c ", each)
		cmd := exec.Command("sh", "-c", each)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("%d: failed to run :%v", i, err)
		}
	}
	return nil
}
