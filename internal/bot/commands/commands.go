package commands

import "github.com/sabinelab/sabine-go/internal/command"

func Get() []command.Command {
	return []command.Command{Ping, Sub, Group, Simple}
}
