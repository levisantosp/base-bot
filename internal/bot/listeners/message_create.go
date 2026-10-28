package listeners

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/sabinelab/sabine-go/internal/command"
)

func MessageCreate(m *command.CommandManager) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		if e.Message.Author.Bot {
			return
		}

		channel, ok := e.Channel()
		if !ok || channel.Type() != discord.ChannelTypeGuildText {
			return
		}

		guild, ok := e.Guild()
		if !ok {
			return
		}

		bot, ok := e.Client().Caches.SelfMember(guild.ID)
		if !ok {
			return
		}

		ctx := command.NewCommandContext(command.CommandContextOpts{
			Client: e.Client(),
			Guild:  &guild,
			Data:   e,
			Locale: "en-us",
			Rest:   e.Client().Rest,
		})

		permissions := e.Client().Caches.MemberPermissionsInChannel(
			channel,
			*e.Message.Member,
		)
		botPermissions := e.Client().Caches.MemberPermissionsInChannel(
			channel,
			bot,
		)

		if err := m.Exec(ctx, permissions, botPermissions); err != nil {
			slog.Error(err.Error())
		}
	})
}
