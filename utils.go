package main

import (
	"errors"
	"log"
)

func printError(args ...interface{}) {
	log.Println(append([]interface{}{"\033[1;31mERROR:\033[0m"}, args...)...)
}

var errAbort = errors.New("gmig aborted")
