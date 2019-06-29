package cmd

import (
	"fmt"
	"strconv"
)

type startCommand struct {
	host string `long:"host" short:"h" description:"Binding host of the server"`
	port int    `long:"port" short:"p" description:"Listening port of the server"`
}

func (cmd *startCommand) Execute(args []string) error {
	fmt.Println("Starting server on host " + cmd.host + " and port " + strconv.Itoa(cmd.port))
	return nil
}

type MainCommand struct {
	Start startCommand `command:"start" description:"Start server on predefined host and port"`
}
