package command

import (
	"sort"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/sabinelab/sabine-go/internal/i18n"
)

type CommandArgument struct {
	Type                     int
	Name                     string
	NameLocalizations        map[discord.Locale]string
	Description              string
	DescriptionLocalizations map[discord.Locale]string
	Required                 any
	Autocomplete             bool
	Choices                  any
	MinValue                 *float64
	MaxValue                 *float64
	MinLength                *int
	MaxLength                *int
	Args                     map[string]CommandArgument
	Value                    any
}

type CommandArguments map[string]CommandArgument

type CreateAutocompleteInteractionOpts struct {
	Interaction discord.AutocompleteInteraction
	T           any
	Client      bot.Client
	Args        []string
}

type CreateMessageComponentInteractionOpts struct {
	Ctx         any
	T           i18n.Type
	Interaction discord.ComponentInteraction
	Client      bot.Client
}

type CreateModalSubmitInteractionOpts struct {
	Ctx         any
	Client      bot.Client
	T           i18n.Type
	Interaction discord.ModalSubmitInteraction
}

type Command struct {
	Name                              string
	Aliases                           []string
	NameLocalizations                 map[discord.Locale]string
	Description                       string
	DescriptionLocalizations          map[discord.Locale]string
	Args                              CommandArguments
	Permissions                       []discord.Permissions
	BotPermissions                    []discord.Permissions
	OnlyDev                           bool
	Ephemeral                         bool
	IsThinking                        bool
	Run                               func(ctx *CommandContext, client bot.Client) (any, error)
	CreateAutocompleteInteraction     func(interaction discord.AutocompleteInteraction, t any, client bot.Client, args []string) (any, error)
	CreateMessageComponentInteraction func(ctx any, t i18n.Type, interaction discord.ComponentInteraction, client bot.Client) (any, error)
	CreateModalSubmitInteraction      func(ctx any, client bot.Client, t i18n.Type, interaction discord.ModalSubmitInteraction) (any, error)
}

func CreateCommand(command Command) Command {
	return command
}

func ParseArguments(args CommandArguments) []discord.ApplicationCommandOption {
	if args == nil {
		return nil
	}

	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		a, b := args[keys[i]], args[keys[j]]

		if Required(a.Required) != Required(b.Required) {
			return Required(a.Required)
		}

		if a.Name != b.Name {
			return a.Name < b.Name
		}

		return keys[i] < keys[j]
	})

	opts := make([]discord.ApplicationCommandOption, 0, len(args))

	for _, key := range keys {
		arg := args[key]
		switch discord.ApplicationCommandOptionType(arg.Type) {
		case discord.ApplicationCommandOptionTypeSubCommand:
			{
				opts = append(opts, discord.ApplicationCommandOptionSubCommand{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Options: ParseArguments(
						CommandArguments(arg.Args),
					),
				})
			}
		case discord.ApplicationCommandOptionTypeSubCommandGroup:
			{
				children := ParseArguments(CommandArguments(arg.Args))
				subcommands := make(
					[]discord.ApplicationCommandOptionSubCommand,
					0,
					len(children),
				)

				for _, child := range children {
					if subcommand, ok := child.(discord.ApplicationCommandOptionSubCommand); ok {
						subcommands = append(subcommands, subcommand)
					}
				}

				opts = append(
					opts,
					discord.ApplicationCommandOptionSubCommandGroup{
						Name:                     arg.Name,
						NameLocalizations:        arg.NameLocalizations,
						Description:              arg.Description,
						DescriptionLocalizations: arg.DescriptionLocalizations,
						Options:                  subcommands,
					},
				)
			}

		case discord.ApplicationCommandOptionTypeString:
			{
				choices, _ := arg.Choices.([]discord.ApplicationCommandOptionChoiceString)

				opts = append(opts, discord.ApplicationCommandOptionString{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
					Choices:                  choices,
					MinLength:                arg.MinLength,
					MaxLength:                arg.MaxLength,
					Autocomplete:             arg.Autocomplete,
				})
			}
		case discord.ApplicationCommandOptionTypeInt:
			{
				choices, _ := arg.Choices.([]discord.ApplicationCommandOptionChoiceInt)

				var minValue, maxValue *int

				if arg.MinValue != nil {
					value := int(*arg.MinValue)
					minValue = &value
				}

				if arg.MaxValue != nil {
					value := int(*arg.MaxValue)
					maxValue = &value
				}

				opts = append(opts, discord.ApplicationCommandOptionInt{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
					Choices:                  choices,
					MinValue:                 minValue,
					MaxValue:                 maxValue,
					Autocomplete:             arg.Autocomplete,
				})
			}

		case discord.ApplicationCommandOptionTypeFloat:
			{
				choices, _ := arg.Choices.([]discord.ApplicationCommandOptionChoiceFloat)

				opts = append(opts, discord.ApplicationCommandOptionFloat{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
					Choices:                  choices,
					MinValue:                 arg.MinValue,
					MaxValue:                 arg.MaxValue,
					Autocomplete:             arg.Autocomplete,
				})
			}

		case discord.ApplicationCommandOptionTypeBool:
			{
				opts = append(opts, discord.ApplicationCommandOptionBool{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}

		case discord.ApplicationCommandOptionTypeUser:
			{
				opts = append(opts, discord.ApplicationCommandOptionUser{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}

		case discord.ApplicationCommandOptionTypeChannel:
			{
				opts = append(opts, discord.ApplicationCommandOptionChannel{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}

		case discord.ApplicationCommandOptionTypeRole:
			{
				opts = append(opts, discord.ApplicationCommandOptionRole{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}

		case discord.ApplicationCommandOptionTypeMentionable:
			{
				opts = append(opts, discord.ApplicationCommandOptionMentionable{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}

		case discord.ApplicationCommandOptionTypeAttachment:
			{
				opts = append(opts, discord.ApplicationCommandOptionAttachment{
					Name:                     arg.Name,
					NameLocalizations:        arg.NameLocalizations,
					Description:              arg.Description,
					DescriptionLocalizations: arg.DescriptionLocalizations,
					Required:                 Required(arg.Required),
				})
			}
		}
	}

	return opts
}
