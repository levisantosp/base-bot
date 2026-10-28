package command

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sabinelab/sabine-go/internal/env"
)

type CommandManager struct {
	Client   *bot.Client
	commands map[string]Command
	aliases  map[string]string
}

func NewCommandManager(client *bot.Client) *CommandManager {
	return &CommandManager{
		Client:   client,
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

func (m *CommandManager) Register(commands []Command) {
	for _, command := range commands {
		m.commands[command.Name] = command

		for _, alias := range command.Aliases {
			m.aliases[alias] = command.Name
		}
	}
}

func (m *CommandManager) get(name string) *Command {
	if command, ok := m.commands[name]; ok {
		return &command
	}

	alias, ok := m.aliases[name]
	if !ok {
		return nil
	}

	if command, ok := m.commands[alias]; ok {
		return &command
	} else {
		return nil
	}
}

func Required(value any) bool {
	switch value := value.(type) {
	case bool:
		return value
	case *bool:
		return value != nil && *value
	case string:
		return value != ""
	default:
		return false
	}
}

func (m *CommandManager) verifyMessageArguments(
	ctx *CommandContext,
	args CommandArguments,
	input []string,
	namespace map[string]any,
) (bool, error) {
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		a, b := args[keys[i]], args[keys[j]]

		if Required(a.Required) != Required(b.Required) {
			return Required(a.Required)
		}

		return a.Name < b.Name
	})

	isSub := false
	for _, key := range keys {
		optType := discord.ApplicationCommandOptionType(args[key].Type)
		if optType == discord.ApplicationCommandOptionTypeSubCommand ||
			optType == discord.ApplicationCommandOptionTypeSubCommandGroup {
			isSub = true
			break
		}
	}

	if isSub {
		if len(input) == 0 {
			_, err := ctx.Reply(
				fmt.Sprintf("Missing argument: `%s`", strings.Join(keys, "/")),
			)
			return false, err
		}

		subName := strings.ToLower(input[0])

		for _, key := range keys {
			arg := args[key]
			if !strings.EqualFold(arg.Name, subName) {
				continue
			}

			nestedNamespace := make(map[string]any)
			namespace[key] = nestedNamespace

			if len(arg.Args) == 0 {
				ctx.Values = nestedNamespace
				return true, nil
			}

			return m.verifyMessageArguments(
				ctx,
				arg.Args,
				input[1:],
				nestedNamespace,
			)
		}

		if len(keys) == 1 {
			key := keys[0]
			arg := args[key]

			nestedNamespace := make(map[string]any)
			namespace[key] = nestedNamespace

			if len(arg.Args) == 0 {
				ctx.Values = nestedNamespace
				return true, nil
			}

			return m.verifyMessageArguments(
				ctx,
				arg.Args,
				input,
				nestedNamespace,
			)
		}

		return true, nil
	}

	for i, key := range keys {
		arg := args[key]

		if i >= len(input) {
			if Required(arg.Required) {
				_, err := ctx.Reply(
					fmt.Sprintf("Missing argument: %s", arg.Name),
				)
				return false, err
			}
			continue
		}

		value, err := m.parseMessageArguments(
			ctx,
			discord.ApplicationCommandOptionType(arg.Type),
			input[i],
		)
		if err != nil {
			_, err := ctx.Reply(
				fmt.Sprintf("Missing argument: %s", arg.Name),
			)
			return false, err
		}

		namespace[key] = value
	}

	ctx.Values = namespace

	return true, nil
}

func (m *CommandManager) parseMessageArguments(
	ctx *CommandContext,
	optType discord.ApplicationCommandOptionType,
	value string,
) (any, error) {
	switch optType {
	case discord.ApplicationCommandOptionTypeString:
		return value, nil
	case discord.ApplicationCommandOptionTypeInt:
		return strconv.Atoi(value)
	case discord.ApplicationCommandOptionTypeFloat:
		return strconv.ParseFloat(value, 64)
	case discord.ApplicationCommandOptionTypeBool:
		return strconv.ParseBool(value)
	case discord.ApplicationCommandOptionTypeUser:
		{
			raw := regexp.MustCompile(`[<@!>]`).ReplaceAllString(value, "")
			id, err := snowflake.Parse(raw)
			if err != nil {
				return nil, err
			}

			if ctx.Guild != nil {
				if member, ok := ctx.Client.Caches.Member(
					ctx.Guild.ID,
					id,
				); ok {
					return member.User, nil
				}
			}

			user, err := ctx.Client.Rest.GetUser(id)
			if err != nil {
				return nil, err
			}
			return *user, nil
		}
	case discord.ApplicationCommandOptionTypeChannel:
		{
			raw := regexp.MustCompile(`[<#>]`).ReplaceAllString(value, "")
			channelID, err := snowflake.Parse(raw)
			if err != nil {
				return nil, err
			}

			if channel, ok := ctx.Client.Caches.Channel(channelID); ok {
				return channel, nil
			}

			channel, err := ctx.Client.Rest.GetChannel(channelID)
			if err != nil {
				return nil, err
			}

			guildChannel, ok := channel.(discord.GuildChannel)
			if !ok {
				return nil, nil
			}

			return guildChannel, nil
		}
	case discord.ApplicationCommandOptionTypeRole:
		{
			raw := regexp.MustCompile(`[<@&>]`).ReplaceAllString(value, "")
			roleID, err := snowflake.Parse(raw)
			if err != nil {
				return nil, err
			}

			if role, ok := ctx.Client.Caches.Role(ctx.Guild.ID, roleID); ok {
				return role, nil
			}

			role, err := ctx.Client.Rest.GetRole(ctx.Guild.ID, roleID)
			if err != nil {
				return nil, err
			}
			return *role, nil
		}
	default:
		return value, nil
	}
}

func getArgument(
	args CommandArguments,
	name string,
) *CommandArgument {
	for _, arg := range args {
		if arg.Name == name {
			return &arg
		}
	}

	return nil
}

func (m *CommandManager) parseSlashArguments(
	data discord.SlashCommandInteractionData,
	args CommandArguments,
) *map[string]any {
	currentArgs := args

	if data.SubCommandGroupName != nil {
		group := getArgument(args, *data.SubCommandGroupName)
		if group == nil {
			return nil
		}

		currentArgs = group.Args
	}

	if data.SubCommandName != nil {
		sub := getArgument(currentArgs, *data.SubCommandName)
		if sub == nil {
			return nil
		}

		currentArgs = sub.Args
	}

	values := make(map[string]any, len(currentArgs))

	for key, arg := range currentArgs {
		optType := discord.ApplicationCommandOptionType(arg.Type)

		if optType == discord.ApplicationCommandOptionTypeSubCommand ||
			optType == discord.ApplicationCommandOptionTypeSubCommandGroup {
			continue
		}

		value, ok := m.parseSlashArgument(data, arg)
		if !ok {
			if Required(arg.Required) {
				return nil
			}
			continue

		}

		values[key] = value
	}

	return &values
}

func (m *CommandManager) parseSlashArgument(
	data discord.SlashCommandInteractionData,
	arg CommandArgument,
) (any, bool) {
	switch discord.ApplicationCommandOptionType(arg.Type) {
	case discord.ApplicationCommandOptionTypeString:
		return data.OptString(arg.Name)
	case discord.ApplicationCommandOptionTypeInt:
		return data.OptInt(arg.Name)
	case discord.ApplicationCommandOptionTypeFloat:
		return data.OptFloat(arg.Name)
	case discord.ApplicationCommandOptionTypeBool:
		return data.OptBool(arg.Name)
	case discord.ApplicationCommandOptionTypeUser:
		return data.OptUser(arg.Name)
	case discord.ApplicationCommandOptionTypeChannel:
		return data.OptChannel(arg.Name)
	case discord.ApplicationCommandOptionTypeRole:
		return data.OptRole(arg.Name)
	case discord.ApplicationCommandOptionTypeMentionable:
		return data.OptMentionable(arg.Name)
	case discord.ApplicationCommandOptionTypeAttachment:
		return data.OptAttachment(arg.Name)
	default:
		return nil, false
	}
}

func (m *CommandManager) Exec(
	ctx *CommandContext,
	permissions discord.Permissions,
	botPermissions discord.Permissions,
) error {
	var commandName string
	args := []string{}

	if i, ok := ctx.Data.(*handler.CommandEvent); ok {
		commandName = i.Data.CommandName()
	} else if msg, ok := ctx.Data.(*events.MessageCreate); ok {
		if !strings.HasPrefix(msg.Message.Content, env.Data.Prefix) {
			return nil
		}

		msgArray := strings.Split(msg.Message.Content, " ")
		command := strings.ToLower(msgArray[0])
		args = append(args, msgArray[1:]...)
		commandName = command[len(env.Data.Prefix):]
	}

	cmd := m.get(commandName)
	if cmd == nil {
		return nil
	}

	if cmd.Permissions != nil &&
		permissions.Missing(cmd.Permissions...) {
		var perms []string
		for _, p := range cmd.BotPermissions {
			perms = append(
				perms,
				fmt.Sprintf("`%s`", ctx.T(fmt.Sprintf("permissions.%s", p))),
			)
		}

		_, err := ctx.Reply(
			fmt.Sprintf(
				"You are missing permissions: %s",
				strings.Join(perms, ", "),
			),
		)
		return err
	}

	if cmd.BotPermissions != nil &&
		botPermissions.Missing(cmd.BotPermissions...) {
		var perms []string
		for _, p := range cmd.BotPermissions {
			perms = append(
				perms,
				fmt.Sprintf("`%s`", ctx.T(fmt.Sprintf("permissions.%s", p))),
			)
		}

		_, err := ctx.Reply(
			fmt.Sprintf(
				"Bot is missing permissions: %s",
				strings.Join(perms, ", "),
			),
		)
		return err
	}

	if cmd.Args != nil {
		if i, ok := ctx.Data.(*handler.CommandEvent); ok {
			values := m.parseSlashArguments(
				i.SlashCommandInteractionData(),
				cmd.Args,
			)
			if values == nil {
				return nil
			}

			ctx.Values = *values
		} else if _, ok := ctx.Data.(*events.MessageCreate); ok {
			namespace := make(map[string]any)
			success, err := m.verifyMessageArguments(
				ctx,
				cmd.Args,
				args,
				namespace,
			)
			if err != nil || !success {
				return err
			}
		}
	}

	i, isSlash := ctx.Data.(*handler.CommandEvent)
	e, isMsg := ctx.Data.(*events.MessageCreate)
	if cmd.Ephemeral && isSlash {
		if err := i.DeferCreateMessage(true); err != nil {
			return err
		}
		ctx.Deferred = true
	} else if cmd.IsThinking {
		if isSlash {
			if err := i.DeferCreateMessage(false); err != nil {
				return err
			}
			ctx.Deferred = true
		} else if isMsg {
			err := ctx.rest.SendTyping(e.ChannelID)
			if err != nil {
				return err
			}
		}
	}

	_, err := cmd.Run(ctx, *m.Client)
	if err == nil {

		if cmd.Ephemeral {
			if e, ok := ctx.Data.(*events.MessageCreate); ok {
				err := ctx.rest.DeleteMessage(e.ChannelID, e.MessageID)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	_, _ = ctx.Reply("An unexpected error has occurred")

	return nil
}
