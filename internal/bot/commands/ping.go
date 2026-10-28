package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/sabinelab/sabine-go/internal/command"
)

var Ping = command.CreateCommand(command.Command{
	Name:        "ping",
	Description: "Shows the bot latency",
	DescriptionLocalizations: map[discord.Locale]string{
		"pt-BR": "Mostra a latência do bot",
	},
	Run: func(ctx *command.CommandContext, client bot.Client) (any, error) {
		shard, err := client.Shard(ctx.Guild.ID)
		if err != nil {
			return nil, err
		}

		return ctx.Reply(
			fmt.Sprintf("Pong! `%dms`", shard.Latency().Milliseconds()),
		)
	},
})
