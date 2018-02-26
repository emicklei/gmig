package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ProjectsIAMPolicy is for capturing the project iam policy
type ProjectsIAMPolicy struct {
	Bindings []struct {
		Members []string
		Role    string
	}
}

// ExportProjectsIAMPolicy reads the current IAM bindings on project level
// and outputs the contents of a gmig migration file.
// Return the filename of the migration.
func ExportProjectsIAMPolicy(cfg Config) error {
	out := new(bytes.Buffer)
	cmdline := []string{"gcloud", "projects", "get-iam-policy", cfg.Project, "--format", "json"}
	if cfg.verbose {
		log.Println(strings.Join(cmdline, " "))
	}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	var p ProjectsIAMPolicy
	if err := json.Unmarshal(out.Bytes(), &p); err != nil {
		return err
	}
	// Build reverse mapping
	memberToRoles := map[string][]string{}
	for _, each := range p.Bindings {
		role := each.Role
		for _, member := range each.Members {
			list, ok := memberToRoles[member]
			if !ok {
				list = []string{}
			}
			memberToRoles[member] = append(list, role)
		}
	}
	content := new(bytes.Buffer)
	fmt.Fprintln(content, "# exported projects iam policy")
	fmt.Fprint(content, "\ndo:")
	for member, roles := range memberToRoles {
		fmt.Fprintf(content, "\n  # member = %s\n", member)
		for _, role := range roles {
			cmd := fmt.Sprintf("  - gcloud projects add-iam-policy-binding $PROJECT --member %s --role %s\n", member, role)
			fmt.Fprintf(content, cmd)
		}
	}
	fmt.Fprintf(content, "\nundo:")
	for member, roles := range memberToRoles {
		fmt.Fprintf(content, "\n  # member = %s\n", member)
		for _, role := range roles {
			cmd := fmt.Sprintf("  - gcloud projects remove-iam-policy-binding $PROJECT --member %s --role %s\n", member, role)
			fmt.Fprintf(content, cmd)
		}
	}
	filename := NewFilename("exported project iam policy")
	if cfg.verbose {
		log.Println("writing", filename)
	}
	return ioutil.WriteFile(filename, content.Bytes(), os.ModePerm)
}
