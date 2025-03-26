package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/urfave/cli"
)

func cmdTemplate(c *cli.Context) error {
	source := c.Args().First()
	isRewrite := c.Bool("w")
	data, err := os.ReadFile(source)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	funcs := template.FuncMap{"env": os.Getenv}
	tmpl, err := template.New("tmpl").Funcs(funcs).Parse(string(data))
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	output := new(bytes.Buffer)
	err = tmpl.Execute(output, "")
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	if isRewrite {
		return ioutil.WriteFile(source, output.Bytes(), os.FileMode(0644)) // -rw-r--r--, see http://permissions-calculator.org/
	}
	fmt.Fprintln(os.Stdout, output)
	return nil
}
