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
	Filename    string `yaml:"-"`
	Description string `yaml:"-"`
	Up          string `yaml:"up"`
	Down        string `yaml:"down"`
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
	// cannot use yaml.Marshal because it does not output JSON in a dev friendly way
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "# %s\n", m.Description)
	// yaml wants spaces as prefix
	fmt.Fprintf(out, "\n# write a shell command here (indent with 2 spaces)\n")
	fmt.Fprintf(out, "up: >\n  %s\n", m.Up)
	// yaml wants spaces as prefix
	fmt.Fprintf(out, "\n# write a shell command here that reverts the effect of up:\n")
	fmt.Fprintf(out, "down: >\n  %s\n", m.Down)
	return out.Bytes(), nil
}

// Execute the commands for this migration.
func (m Migration) Execute(command string) error {
	if len(command) == 0 {
		return nil
	}
	log.Println("sh -c ", command)
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run :%v", err)
	}
	return nil
}
