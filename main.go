package main

import (
	"github.com/jessevdk/go-flags"

	"fmt"
	"os"
)

type GreetCommand struct {
	Message string `long:"message"`
}

func (cmd *GreetCommand) Execute(args []string) error {
	fmt.Println(cmd.Message)
	return nil
}

type MainCommand struct {
	Greet GreetCommand `command:"greet" description:"Send polite greeting message"`
}

func main() {
	var mainCmd MainCommand

	parser := flags.NewParser(&mainCmd, flags.HelpFlag|flags.PassDoubleDash)
	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
