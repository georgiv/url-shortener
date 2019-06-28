package cmd

import (
	"fmt"
)

type StartCommand struct {
	host string
	port int
}

func (*StartCommand) Execute(args []string) {
	fmt.Println("Hello, Go World! :)")
}
