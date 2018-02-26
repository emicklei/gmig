package main

import "os/exec"

type commandCapturer struct {
	args   [][]string
	output []byte
}

func (c *commandCapturer) runCommand(cmd *exec.Cmd) ([]byte, error) {
	c.args = append(c.args, cmd.Args)
	return c.output, nil
}
