package cmd

// MainCommand represents all supported commands
type MainCommand struct {
	Start StartCommand `command:"start" description:"Start server on predefined host and port"`
}
