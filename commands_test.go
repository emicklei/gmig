package main

import (
	"os"
	"testing"
)

func TestCmdInit(t *testing.T) {
	os.Chdir("test")
	os.Args = []string{"gmig", "init", "demo"}
	main()
}
