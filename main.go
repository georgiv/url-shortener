package main

import (
	"log"
	"os"

	"github.com/georgiv/url-shortener/server/cmd"
	"github.com/jessevdk/go-flags"
)

func main() {
	var mainCmd cmd.MainCommand

	parser := flags.NewParser(&mainCmd, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
