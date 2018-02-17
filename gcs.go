package gmig

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"

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
	capturedErr := new(bytes.Buffer)
	cmd := exec.Command("gsutil", "-q", "cp",
		filepath.Join(g.Configuration.Bucket, g.Configuration.StateObject),
		localStateFilename)
	cmd.Stderr = capturedErr
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if len(capturedErr.String()) > 0 {
		return "", errors.New("stderr:" + capturedErr.String())
	}
	return g.onDiskAccess.LoadState()
}

func (g GCS) SaveState(filename string) error {
	if err := g.onDiskAccess.SaveState(filename); err != nil {
		return err
	}
	capturedErr := new(bytes.Buffer)
	cmd := exec.Command("gsutil", "-q", "cp",
		localStateFilename,
		filepath.Join(g.Configuration.Bucket, g.Configuration.StateObject))
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
