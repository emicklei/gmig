package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/emicklei/tre"
)

func gcloudConfigList() {
	log.Println("checking gcloud config list ...")
	cmd := exec.Command("gcloud", "config", "list")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
}

func gcloudConfigSetProject(cfg Config) error {
	for _, each := range []struct {
		Key, Value string
	}{
		{"project", cfg.Project},
		{"region", cfg.Region},
		{"zone", cfg.Zone},
	} {
		k := each.Key
		v := each.Value
		if len(v) > 0 { // skip optional values
			if cfg.verbose {
				log.Printf("setting gcloud config [%s] to [%s]\n", k, v)
			}
			cmd := exec.Command("gcloud", "config", "set", k, v)
			data, err := runCommand(cmd)
			if cfg.verbose {
				log.Println(string(data))
			}
			if err != nil {
				return tre.New(err, "error changing gcloud config", k, v)
			}
		}
	}
	return nil
}
