package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ExportStorageIAMPolicy prints a migration that sets IAM for each bucket owned by the project.
func ExportStorageIAMPolicy(cfg Config) error {
	// get all buckets
	cmdline := []string{"gsutil", "list"}
	if cfg.verbose {
		log.Println(strings.Join(cmdline, " "))
	}
	out := new(bytes.Buffer)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	buckets := strings.Split(strings.TrimSpace(out.String()), "\n")

	// build data structure for migration
	type memberRolesPerBucket struct {
		bucket        string
		memberToRoles map[string][]string
	}
	list := []memberRolesPerBucket{}
	for _, each := range buckets {
		policy, err := fetchIAMPolicy([]string{"gsutil", "iam", "get", each}, cfg.verbose)
		if err != nil {
			return err
		}
		list = append(list, memberRolesPerBucket{
			bucket:        each,
			memberToRoles: policy.buildMemberToRoles(),
		})
	}

	// begin writing migration
	content := new(bytes.Buffer)
	fmt.Fprintln(content, "# exported buckets iam policy")
	fmt.Fprint(content, "\ndo:")
	for _, each := range list {
		fmt.Fprintf(content, "\n  # bucket = %s\n", each.bucket)
		if len(each.memberToRoles) == 3 { // projectViewer,Editor,Owner // skip defaults
			continue
		}
		for member, roles := range each.memberToRoles {
			if strings.HasPrefix(member, "project") { // skip defaults
				continue
			}
			fmt.Fprintf(content, "\n  # member = %s\n", member)
			for _, role := range roles {
				shortRole := strings.TrimPrefix(role, "roles/storage.")
				cmd := fmt.Sprintf("  - gsutil iam ch %s:%s %s\n", member, shortRole, each.bucket)
				fmt.Fprintln(content, cmd)
			}
		}
	}

	// UNDO section
	fmt.Fprint(content, "\nundo:")
	for _, each := range list {
		fmt.Fprintf(content, "\n  # bucket = %s\n", each.bucket)
		for member, roles := range each.memberToRoles {
			if strings.HasPrefix(member, "project") { // skip defaults
				continue
			}
			fmt.Fprintf(content, "\n  # member = %s\n", member)
			for _, role := range roles {
				shortRole := strings.TrimPrefix(role, "roles/storage.")
				cmd := fmt.Sprintf("  - gsutil iam ch -d %s:%s %s\n", member, shortRole, each.bucket)
				fmt.Fprint(content, cmd)
			}
		}
	}

	// write the migration
	filename := NewFilenameWithIndex("exported buckets iam policy")
	if cfg.verbose {
		log.Println("writing", filename)
	}
	return ioutil.WriteFile(filename, content.Bytes(), os.ModePerm)
}
