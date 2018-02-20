package main

import (
	"fmt"
	"log"
	"os/exec"
)

func gcloudConfigList() {
	log.Println("checking gcloud config list ...")
	cmd := exec.Command("gcloud", "config", "list")
	out, _ := cmd.CombinedOutput()
	fmt.Println(string(out))
}

func gcloudConfigSetProject(project string) {
	log.Printf("setting gcloud config project to [%s]\n", project)
	cmd := exec.Command("gcloud", "config", "set", "project", project)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln(err)
	}
}
