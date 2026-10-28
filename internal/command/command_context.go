package command

import (
	"errors"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"
	"github.com/sabinelab/sabine-go/internal/i18n"
)

type CommandContextOpts struct {
	Client    *bot.Client
	Guild     *discord.Guild
	Data      any
	Locale    string
	Args      CommandArguments
	Values    map[string]any
	Author    discord.User
	Prefix    string
	Rest      rest.Rest
	Responded bool
	Deferred  bool
}

type CommandContext struct {
	Client    *bot.Client
	Guild     *discord.Guild
	Data      any
	Locale    string
	Args      CommandArguments
	Values    map[string]any
	Author    discord.User
	rest      rest.Rest
	Responded bool
	Deferred  bool
}

func NewCommandContext(opts CommandContextOpts) *CommandContext {
	return &CommandContext{
		Client: opts.Client,
		Guild:  opts.Guild,
		Data:   opts.Data,
		Locale: opts.Locale,
		Args:   opts.Args,
		Values: opts.Values,
		Author: opts.Author,
		rest:   opts.Rest,
	}
}

func (ctx *CommandContext) getValue(path ...string) (any, bool) {
	var current any = ctx.Values

	for _, part := range path {
		values, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		current, ok = values[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func (ctx *CommandContext) GetOptString(name string) (string, bool) {
	value, ok := ctx.getValue(name)
	if !ok {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

func (ctx *CommandContext) GetString(name string) string {
	value, _ := ctx.GetOptString(name)
	return value
}

func (ctx *CommandContext) T(content string, args ...i18n.Args) string {
	return i18n.T(ctx.Locale, content, args...)
}

func (ctx *CommandContext) Reply(
	content string,
	args ...i18n.Args,
) (*discord.Message, error) {
	message := discord.MessageCreate{
		Content: ctx.T(content, args...),
	}

	switch data := ctx.Data.(type) {
	case *handler.CommandEvent:
		{
			if ctx.Deferred {
				content := message.Content
				response, err := data.UpdateInteractionResponse(
					discord.MessageUpdate{
						Content: &content,
					},
				)
				if err != nil {
					return nil, err
				}

				ctx.Responded = true
				ctx.Deferred = false

				return response, nil
			}

			if ctx.Responded {
				response, err := data.CreateFollowupMessage(message)
				if err != nil {
					return nil, err
				}

				ctx.Responded = true

				return response, nil
			}

			if err := data.CreateMessage(message); err != nil {
				return nil, err
			}

			ctx.Responded = true
			ctx.Deferred = false

			return data.GetInteractionResponse()
		}
	case *events.MessageCreate:
		{
			return ctx.rest.CreateMessage(
				data.ChannelID,
				message.WithMessageReferenceByID(data.MessageID),
			)
		}
	default:
		{
			return nil, errors.New("unsupported command context data type")
		}
	}
}

func (ctx *CommandContext) ReplyOpts(
	content discord.MessageCreate,
) (*discord.Message, error) {
	switch data := ctx.Data.(type) {
	case *handler.CommandEvent:
		{
			if ctx.Deferred {
				message, err := data.UpdateInteractionResponse(
					discord.MessageUpdate{
						Content: &content.Content,
					},
				)
				if err != nil {
					return nil, err
				}

				ctx.Responded = true
				ctx.Deferred = false
				return message, nil
			}

			if ctx.Responded {
				message, err := data.CreateFollowupMessage(content)
				if err != nil {
					return nil, err
				}

				ctx.Responded = true
				return message, nil
			}

			if err := data.CreateMessage(content); err != nil {
				return nil, err
			}

			ctx.Responded = true
			ctx.Deferred = false

			return data.GetInteractionResponse()
		}
	case *events.MessageCreate:
		{
			return ctx.rest.CreateMessage(
				data.ChannelID,
				content.WithMessageReferenceByID(data.MessageID),
			)
		}
	default:
		{
			return nil, errors.New("unsupported command context data")
		}
	}
}

func (ctx *CommandContext) Edit(
	content string,
	args ...i18n.Args,
) (*discord.Message, error) {
	t := ctx.T(content, args...)
	message := discord.MessageUpdate{
		Content: &t,
	}

	switch data := ctx.Data.(type) {
	case *handler.CommandEvent:
		{
			message, err := data.UpdateInteractionResponse(message)
			if err != nil {
				return nil, err
			}

			ctx.Responded = true
			ctx.Deferred = false

			return message, nil
		}
	case *events.MessageCreate:
		{
			return ctx.rest.UpdateMessage(
				data.ChannelID,
				data.MessageID,
				message,
			)
		}
	default:
		{
			return nil, errors.New("unsupported command context data")
		}
	}
}

func (ctx *CommandContext) EditOpts(
	content discord.MessageUpdate,
	args ...i18n.Args,
) (*discord.Message, error) {
	switch data := ctx.Data.(type) {
	case *handler.CommandEvent:
		{
			message, err := data.UpdateInteractionResponse(content)
			if err != nil {
				return nil, err
			}

			ctx.Responded = true
			ctx.Deferred = false

			return message, nil
		}
	case *events.MessageCreate:
		{
			return ctx.rest.UpdateMessage(
				data.ChannelID,
				data.MessageID,
				content,
			)
		}
	default:
		{
			return nil, errors.New("unsupported command context data")
		}
	}
}
