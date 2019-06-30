package cmd

import (
	"github.com/georgiv/url-shortener/server/web"
)

type StartCommand struct {
	Host string `long:"host" short:"h" default:"localhost" description:"Host where to bind the server"`
	Port int    `long:"port" short:"p" default:"8888" description:"Listening port of the server"`
}

func (cmd *StartCommand) Execute(args []string) error {
	s := web.NewServer(cmd.Host, cmd.Port)
	s.Handle()
	return nil
}
