package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/urfave/cli"
)

func cmdTemplate(c *cli.Context) error {
	source := c.Args().First()
	isRewrite := c.Bool("w")
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}
	funcs := template.FuncMap{"env": os.Getenv}
	tmpl, err := template.New("tmpl").Funcs(funcs).Parse(string(data))
	if err != nil {
		log.Fatal(err)
	}
	output := new(bytes.Buffer)
	err = tmpl.Execute(output, "")
	if err != nil {
		log.Fatal(err)
	}
	if isRewrite {
		return ioutil.WriteFile(source, output.Bytes(), os.FileMode(0644)) // -rw-r--r--, see http://permissions-calculator.org/
	}
	fmt.Fprintln(os.Stdout, output)
	return nil
}
