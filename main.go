package main

import (
	"github.com/cranki/url-shortener/server/cmd/start"
	"github.com/jessevdk/go-flags"

	"fmt"
	"os"
)

func main() {
	var mainCmd start.MainCommand

	parser := flags.NewParser(&mainCmd, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
