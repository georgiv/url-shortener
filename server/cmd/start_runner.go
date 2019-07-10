package cmd

import (
	"fmt"

	"github.com/georgiv/url-shortener/server/web"
)

type StartCommand struct {
	Host       string `long:"bindhost" short:"b" default:"localhost" description:"Host where to bind the server"`
	Port       int    `long:"port" short:"p" default:"8888" description:"Listening port of the server"`
	Expiration int    `long:"expiration" short:"e" default:"7" description:"Expiration time for short urls in days"`
}

func (cmd *StartCommand) Execute(args []string) error {
	s, err := web.NewServer(cmd.Host, cmd.Port, cmd.Expiration)
	if err != nil {
		return fmt.Errorf("Error while starting web server: %v", err)
	}

	s.Handle()

	return nil
}
