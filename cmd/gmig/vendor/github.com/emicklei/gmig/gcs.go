package gmig

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/emicklei/tre"
)

const localStateFilename = ".gmig.state"

// GCS = Google Cloud Storage
type GCS struct {
	Configuration Config
	onDiskAccess  FileStateProvider
}

// LoadState implements StateProvider
func (g GCS) LoadState() (string, error) {
	cmdline := []string{"gsutil", "-q", "cp",
		"gs://" + filepath.Join(g.Configuration.Bucket, g.Configuration.StateObject),
		localStateFilename}
	if err := g.gsutil(cmdline); err != nil {
		return "", err
	}
	return g.onDiskAccess.LoadState()
}

// SaveState implements StateProvider
func (g GCS) SaveState(filename string) error {
	if err := g.onDiskAccess.SaveState(filename); err != nil {
		return err
	}
	cmdline := []string{"gsutil", "-q", "cp",
		localStateFilename,
		"gs://" + filepath.Join(g.Configuration.Bucket, g.Configuration.StateObject)}
	return g.gsutil(cmdline)
}

func (g GCS) gsutil(cmdline []string) error {
	if g.Configuration.Verbose {
		log.Println(strings.Join(cmdline, " "))
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	capturedErr := new(bytes.Buffer)
	if g.Configuration.Verbose {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = capturedErr
	err := cmd.Run()
	if err != nil {
		return tre.New(err, "gsutil cp failed")
	}
	if len(capturedErr.String()) > 0 {
		return errors.New("stderr:" + capturedErr.String())
	}
	return nil
}
