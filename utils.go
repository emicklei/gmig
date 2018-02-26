package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/emicklei/tre"
)

func printError(args ...interface{}) {
	log.Println(append([]interface{}{"\033[1;31mERROR:\033[0m"}, args...)...)
}

var errAbort = errors.New("gmig aborted")

func checkExists(filename string) error {
	_, err := os.Stat(filename)
	abs, _ := filepath.Abs(filename)
	return tre.New(err, "no such migration (wrong project?)", "file", abs)
}

// runCommand is wrapper for CombinedOutput to make this package easy testable.
var runCommand = func(c *exec.Cmd) ([]byte, error) {
	return c.CombinedOutput()
}

func reportError(cfg Config, action string, err error) error {
	log.Printf("executing [%s] failed, see error above and or below.\n", action)

	log.Println("checking gmig config ...")
	fmt.Println(cfg.ToJSON())
	gcloudConfigList()
	return err
}
