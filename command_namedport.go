package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

type namedPort struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

const (
	create = 1
	delete = 2
)

func cmdCreateNamedPort(c *cli.Context) error {
	return cmdChangeNamedPort(c, create)
}

func cmdDeleteNamedPort(c *cli.Context) error {
	return cmdChangeNamedPort(c, delete)
}

func cmdChangeNamedPort(c *cli.Context, action int) error {
	if len(c.Args()) != 2 {
		return fmt.Errorf("missing command argument, expected [INSTANCE_GROUP] NAME:PORT")
	}
	verbose := c.GlobalBool("v")
	instanceGroup := c.Args()[0]
	if len(instanceGroup) == 0 {
		return fmt.Errorf("missing Compute Instance Group command argument")
	}
	parts := strings.Split(c.Args()[1], ":")
	if len(parts) == 0 {
		return fmt.Errorf("invalid NAME:PORT command argument")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("PORT must be a positive integer:%v", err)
	}
	name := parts[0]
	// get named ports
	args := []string{"compute", "instance-groups", "get-named-ports", instanceGroup, "--format", "json"}
	cmd := exec.Command("gcloud", args...)
	if verbose {
		log.Println(strings.Join(append([]string{"gcloud"}, args...), " "))
	}
	data, err := runCommand(cmd)
	if err != nil {
		log.Println(string(data)) // stderr
		return fmt.Errorf("failed to run gcloud get-named-ports: %v", err)
	}
	// parse named ports
	var list []namedPort
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&list); err != nil {
		return fmt.Errorf("parsing JSON output failed:%v", err)
	}
	if create == action {
		// only append if not exists, update otherwise or abort
		updated := false
		for _, each := range list {
			if each.Name == name {
				if each.Port == port {
					log.Printf("named-port [%s:%d] already exists\n", name, port)
					return nil
				}
				each.Port = port
				updated = true
				break
			}
		}
		if !updated {
			list = append(list, namedPort{Name: name, Port: port})
		}
	}
	if delete == action {
		// only delete if exists, update otherwise
		deleted := false
		copyWithout := []namedPort{}
		for _, each := range list {
			if each.Name == name {
				deleted = true
			} else {
				copyWithout = append(copyWithout, each)
			}
		}
		if !deleted {
			log.Printf("named-port [%s:%d] did not exists\n", name, port)
			return nil
		}
		list = copyWithout
	}
	// composed set argument
	newList := new(strings.Builder)
	for i, each := range list {
		if i > 0 {
			newList.WriteString(",")
		}
		fmt.Fprintf(newList, "%s:%d", each.Name, each.Port)
	}
	// set named ports using new list
	args = []string{"compute", "instance-groups", "set-named-ports", instanceGroup, "--named-ports", newList.String()}
	cmd = exec.Command("gcloud", args...)
	if verbose {
		log.Println(strings.Join(append([]string{"gcloud"}, args...), " "))
	}
	if data, err := runCommand(cmd); err != nil {
		log.Println(string(data)) // stderr
		return fmt.Errorf("failed to run gcloud set-named-ports: %v", err)
	}
	return nil
}
