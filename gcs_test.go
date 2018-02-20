package main

import (
	"os"
	"testing"
)

func TestLoadStateFromGCS(t *testing.T) {
	bucket := os.Getenv("BB")
	if len(bucket) == 0 {
		t.Log("set BB environment variable to a valid accessible Google Storaget Bucket name (without the gs:// prefix)")
		t.Skip()
	}
	gcs := NewGCS(Config{Bucket: bucket, verbose: true, LastMigrationObjectName: "stateobject"})
	t.Log("save state")
	if err := gcs.SaveState("temp"); err != nil {
		t.Fatal(err)
	}
	t.Log("load state")
	_, err := gcs.LoadState()
	if err != nil {
		t.Error(err)
	}
}
