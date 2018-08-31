package main

import (
	"os/exec"
	"testing"
)

type commandCapturer struct {
	args   [][]string
	output []byte
}

func (c *commandCapturer) runCommand(cmd *exec.Cmd) ([]byte, error) {
	c.args = append(c.args, cmd.Args)
	return c.output, nil
}

func TestPrettyPrint(t *testing.T) {
	testStr := "20180227t140600_permit_infra_manager_to_deploy_to_gateway_cluster.yaml"
	expected := "2018-02-27 14:06:00 permit infra manager to deploy to gateway cluster (20180227t140600_permit_infra_manager_to_deploy_to_gateway_cluster.yaml)"

	if got, want := pretty(testStr), expected; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
