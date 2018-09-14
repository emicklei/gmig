package main

import (
	"os/exec"
	"testing"
)

type commandCapturer struct {
	args   [][]string
	output []byte
	err    error
}

func (c *commandCapturer) runCommand(cmd *exec.Cmd) ([]byte, error) {
	c.args = append(c.args, cmd.Args)
	return c.output, c.err
}

var dateTests = []struct {
	in  string
	out string
}{
	{"20180227t140600_permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "2018-02-27 14:06:00 permit infra manager to deploy to gateway cluster"},
	{"20161230t235959_permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "2016-12-30 23:59:59 permit infra manager to deploy to gateway cluster"},
	{"20170604t031224_permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "2017-06-04 03:12:24 permit infra manager to deploy to gateway cluster"},
	{"20180101t000000_permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "2018-01-01 00:00:00 permit infra manager to deploy to gateway cluster"},
	{"20011122t090620_permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "2001-11-22 09:06:20 permit infra manager to deploy to gateway cluster"},
	{"permit_infra_manager_to_deploy_to_gateway_cluster.yaml", "permit infra manager to deploy to gateway cluster"},
	{"permit.yaml", "permit"},
}

func TestPrettyPrint(t *testing.T) {
	for _, tt := range dateTests {
		actual := pretty(tt.in)
		if actual != tt.out {
			t.Errorf("pretty(%v): expected %v, actual %v", tt.in, tt.out, actual)
		}
	}
}
