package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/sabinelab/sabine-go/internal/bot/commands"
	"github.com/sabinelab/sabine-go/internal/bot/listeners"
	"github.com/sabinelab/sabine-go/internal/command"
	"github.com/sabinelab/sabine-go/internal/env"
)

var cmds = []discord.ApplicationCommandCreate{}

func main() {
	env.LoadEnv()

	shutdown, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	h := handler.New()

	client, err := disgo.New(
		env.Data.BotToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),
		bot.WithEventListeners(h),
		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.FlagGuilds,
				cache.FlagMembers,
				cache.FlagChannels,
			),
		),
	)
	if err != nil {
		panic(err)
	}

	m := command.NewCommandManager(client)

	m.Register(commands.Get())

	client.EventManager.AddEventListeners(
		listeners.MessageCreate(m),
		listeners.Ready(),
	)

	for _, cmd := range commands.Get() {
		cmds = append(cmds, discord.SlashCommandCreate{
			Name:                     cmd.Name,
			NameLocalizations:        cmd.NameLocalizations,
			Description:              cmd.Description,
			DescriptionLocalizations: cmd.DescriptionLocalizations,
			Options:                  command.ParseArguments(cmd.Args),
		})

		h.SlashCommand(
			"/"+cmd.Name,
			func(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
				values := make(map[string]any, len(data.Options))
				for name, option := range data.Options {
					if option.Type == discord.ApplicationCommandOptionTypeString {
						values[name] = option.String()
					}
				}

				guild, ok := e.Guild()
				if !ok {
					return errors.New("guild is not 'ok'")
				}

				var permissions discord.Permissions
				if member := e.Member(); member != nil {
					permissions = member.Permissions
				}

				var botPermissions discord.Permissions
				if p := e.AppPermissions(); p != nil {
					botPermissions = *p
				}

				ctx := command.NewCommandContext(command.CommandContextOpts{
					Client: client,
					Guild:  &guild,
					Data:   e,
					Locale: "en-us",
					Args:   cmd.Args,
					Values: values,
					Author: e.User(),
					Prefix: ",",
					Rest:   client.Rest,
				})

				return m.Exec(ctx, permissions, botPermissions)
			},
		)
	}

	if err := handler.SyncCommands(client, cmds, nil); err != nil {
		panic(err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = client.OpenGateway(ctx); err != nil {
		panic(err)
	}

	<-shutdown.Done()
}
