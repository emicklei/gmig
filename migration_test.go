package gmig

import (
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
