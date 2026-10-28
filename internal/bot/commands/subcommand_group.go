package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/sabinelab/sabine-go/internal/command"
)

var Group = command.CreateCommand(command.Command{
	Name:        "group",
	Description: "Command with a subcommand group",
	Args: command.CommandArguments{
		"group": {
			Type: int(
				discord.ApplicationCommandOptionTypeSubCommandGroup,
			),
			Name:        "group",
			Description: "Subcommand group",
			Args: map[string]command.CommandArgument{
				"subcommand": {
					Type: int(
						discord.ApplicationCommandOptionTypeSubCommand,
					),
					Name:        "subcommand",
					Description: "Subcommand with options",
					Args: map[string]command.CommandArgument{
						"opt1": {
							Type: int(
								discord.ApplicationCommandOptionTypeString,
							),
							Name:        "opt1",
							Description: "Required option",
							Required:    true,
						},
						"opt2": {
							Type: int(
								discord.ApplicationCommandOptionTypeString,
							),
							Name:        "opt2",
							Description: "Optional option",
						},
					},
				},
			},
		},
	},
	Run: func(ctx *command.CommandContext, _ bot.Client) (any, error) {
		opt1 := ctx.GetString("opt1")
		opt2, ok := ctx.GetOptString("opt2")
		if !ok {
			opt2 = "(not provided)"
		}

		return ctx.Reply(fmt.Sprintf("opt1: %s\nopt2: %s", opt1, opt2))
	},
})
