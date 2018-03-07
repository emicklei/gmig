package main

import (
	"os"
	"testing"
)

func TestExportProjectsIAMPolicy(t *testing.T) {
	project := os.Getenv("PP")
	if len(project) == 0 {
		t.Log("set PP environment variable to a valid accessible Google GCP project")
		t.Skip()
	}
	e := ExportProjectsIAMPolicy(Config{})
	if e != nil {
		t.Fatal(e)
	}
}
