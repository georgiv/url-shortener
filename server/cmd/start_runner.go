package cmd

import (
	"fmt"
	"strconv"
)

type StartCommand struct {
	Host string `long:"host" short:"h" description:"Host where to bind the server"`
	Port int    `long:"port" short:"p" description:"Listening port of the server"`
}

func (cmd *StartCommand) Execute(args []string) error {
	fmt.Println("Starting server on host " + cmd.Host + " and port " + strconv.Itoa(cmd.Port))
	return nil
}
