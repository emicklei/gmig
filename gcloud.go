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

func gcloudConfigSetProject(cfg Config, project string) error {
	if cfg.verbose {
		log.Printf("setting gcloud config project to [%s]\n", project)
	}
	cmd := exec.Command("gcloud", "config", "set", "project", project)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return tre.New(err, "error changing gcloud project")
	}
	return nil
}
