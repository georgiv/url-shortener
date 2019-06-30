package cmd

type MainCommand struct {
	Start StartCommand `command:"start" description:"Start server on predefined host and port"`
}
