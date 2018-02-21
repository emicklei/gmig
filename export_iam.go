package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
func ExportProjectsIAMPolicy(project string) error {
	out := new(bytes.Buffer)
	cmd := exec.Command("gcloud", "projects", "get-iam-policy", project, "--format", "json")
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	var p ProjectsIAMPolicy
	if err := json.Unmarshal(out.Bytes(), &p); err != nil {
		return err
	}
	content := new(bytes.Buffer)
	fmt.Fprintln(content, "# exported projects iam policy")
	fmt.Fprint(content, "\ndo:")
	for _, each := range p.Bindings {
		role := each.Role
		fmt.Fprintf(content, "\n  # role = %s\n", role)
		for _, member := range each.Members {
			cmd := fmt.Sprintf("  - gcloud projects add-iam-policy-binding $PROJECT --member %s --role %s", member, role)
			fmt.Fprintf(content, cmd)
		}
	}
	fmt.Fprintf(content, "\nundo:")
	for _, each := range p.Bindings {
		role := each.Role
		fmt.Fprintf(content, "\n  # role = %s\n", role)
		for _, member := range each.Members {
			cmd := fmt.Sprintf("  - gcloud projects remove-iam-policy-binding $PROJECT --member %s --role %s", member, role)
			fmt.Fprintf(content, cmd)
		}
	}
	filename := NewFilename("exported project iam policy")
	return ioutil.WriteFile(filename, content.Bytes(), os.ModePerm)
}
