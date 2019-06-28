package cmd

import (
	"fmt"
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
