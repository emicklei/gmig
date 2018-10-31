package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func printError(args ...interface{}) {
	log.Println(append([]interface{}{"\033[1;31mERROR:\033[0m"}, args...)...)
}

func printWarning(args ...interface{}) {
	log.Println(append([]interface{}{"\033[1;31mWARNING:\033[0m"}, args...)...)
}

var errAbort = errors.New("gmig aborted")

func checkExists(filename string) error {
	_, err := os.Stat(filename)
	if err == nil {
		return nil
	}
	abs, _ := filepath.Abs(filename)
	return fmt.Errorf("no such migration (wrong project?, git pull?):%s", abs)
}

// runCommand is wrapper for CombinedOutput to make this package easy testable.
var runCommand = func(c *exec.Cmd) ([]byte, error) {
	return c.CombinedOutput()
}

func reportError(cfg Config, action string, err error) error {
	log.Printf("executing [%s] failed, error: [%v]\n", action, err)

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

func promptForYes(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	yn, _ := reader.ReadString('\n')
	return strings.HasPrefix(yn, "Y") || strings.HasPrefix(yn, "y")
}

// 20180227t140600_permit_infra_manager_to_deploy_to_gateway_cluster.yaml
// 2018-02-27 14:06:00 permit infra manager to deploy to gateway cluster
func pretty(filename string) string {

	fn := strings.Replace(strings.TrimSuffix(filename, filepath.Ext(filename)), "_", " ", -1)

	if len(fn) < 16 {
		return fn
	}

	// 20060102t150405 is used as a sample format, see https://golang.org/pkg/time/#Parse
	_, err := time.Parse("20060102t150405", fn[0:15])
	if err != nil {
		return fn
	}

	return fmt.Sprintf("%s-%s-%s %s:%s:%s %s",
		filename[0:4],
		filename[4:6],
		filename[6:8],
		filename[9:11],
		filename[11:13],
		filename[13:15],
		fn[16:])
}

func isYamlFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml"
}

var regexpIndex, _ = regexp.Compile("^[0-9]{3}_")
var regexpTimestamp, _ = regexp.Compile("^[0-9]{8}t[0-9]{6}_")
