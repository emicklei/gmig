package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/emicklei/tre"
	"github.com/urfave/cli"
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

func started(c *cli.Context, action string) func() {
	v := c.GlobalBool("v")
	if !v {
		return func() {}
	}
	log.Println("gmig version", version)
	log.Println("BEGIN", action)
	start := time.Now()
	return func() { log.Println("END", action, "completed in", time.Now().Sub(start)) }
}
