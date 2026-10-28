import {
  ApplicationCommandOptionType,
  type ApplicationCommandOptionChoiceData,
  type ApplicationCommandOptionData,
  type Attachment,
  type AutocompleteInteraction,
  type Channel,
  type LocalizationMap,
  type MessageComponentInteraction,
  type ModalSubmitInteraction,
  type PermissionResolvable,
  type Role,
  type User
} from 'discord.js'
import type { Client } from '../client/Client'
import type { CommandContext } from './CommandContext'

type CommandArgumentType = {
  [ApplicationCommandOptionType.Subcommand]: null
  [ApplicationCommandOptionType.SubcommandGroup]: null
  [ApplicationCommandOptionType.String]: string
  [ApplicationCommandOptionType.Integer]: number
  [ApplicationCommandOptionType.Number]: number
  [ApplicationCommandOptionType.Boolean]: boolean
  [ApplicationCommandOptionType.User]: User
  [ApplicationCommandOptionType.Channel]: Channel
  [ApplicationCommandOptionType.Role]: Role
  [ApplicationCommandOptionType.Mentionable]: User | Role
  [ApplicationCommandOptionType.Attachment]: Attachment
}

type CommandArgumentChoice = ApplicationCommandOptionChoiceData<string | number>

type CommandArgumentValue<
  T extends readonly CommandArgumentChoice[] | undefined
> = T extends readonly (infer Choice)[]
  ? Choice extends { value: infer Value }
    ? Value
    : never
  : never

type CommandArgument = {
  type: keyof CommandArgumentType
  name: string
  nameLocalizations?: LocalizationMap
  description: string
  descriptionLocalizations?: LocalizationMap
  required?: boolean
  autocomplete?: boolean
  choices?: readonly CommandArgumentChoice[]
  min_value?: number
  max_value?: number
  min_length?: number
  max_length?: number
  args?: CommandArguments
}

export type CommandArguments = Record<string, CommandArgument>

export type ResolveArguments<T extends CommandArguments> = {
  [K in keyof T]: T[K]['type'] extends
    | ApplicationCommandOptionType.Subcommand
    | ApplicationCommandOptionType.SubcommandGroup
    ?
        | (T[K]['args'] extends CommandArguments
            ? ResolveArguments<T[K]['args']>
            : Record<string, never>)
        | undefined
    : T[K]['type'] extends
          | ApplicationCommandOptionType.String
          | ApplicationCommandOptionType.Integer
          | ApplicationCommandOptionType.Number
      ? T[K]['choices'] extends readonly CommandArgumentChoice[]
        ? T[K]['required'] extends true
          ? CommandArgumentValue<T[K]['choices']>
          : CommandArgumentValue<T[K]['choices']> | undefined
        : T[K]['required'] extends true
          ? CommandArgumentType[T[K]['type']]
          : CommandArgumentType[T[K]['type']] | undefined
      : T[K]['required'] extends true
        ? CommandArgumentType[T[K]['type']]
        : CommandArgumentType[T[K]['type']] | undefined
}

type CommandOptions<T extends CommandArguments> = {
  ctx: CommandContext<ResolveArguments<T>>
  client: Client
  id: string
}

type CreateAutocompleteInteractionOptions = {
  i: AutocompleteInteraction
  client: Client
  args?: string[]
}

type CreateComponentInteractionOptions = {
  ctx: unknown
  i: MessageComponentInteraction
  client: Client
}

type CreateModalSubmitInteractionOptions = {
  ctx: unknown
  client: Client
  i: ModalSubmitInteraction
}

export type Command<T extends CommandArguments = CommandArguments> = {
  name: string
  aliases?: string[]
  nameLocalizations?: LocalizationMap
  description: string
  descriptionLocalizations?: LocalizationMap
  args?: T
  permissions?: PermissionResolvable[]
  botPermissions?: PermissionResolvable[]
  onlyDev?: boolean
  ephemeral?: boolean
  isThinking?: boolean
  run: (props: CommandOptions<T>) => Promise<unknown>
  createAutocompleteInteraction?: (
    options: CreateAutocompleteInteractionOptions
  ) => Promise<unknown>
  createMessageComponentInteraction?: (
    options: CreateComponentInteractionOptions
  ) => Promise<unknown>
  createModalSubmitInteraction?: (
    options: CreateModalSubmitInteractionOptions
  ) => Promise<unknown>
}

export function parseArguments(
  args?: CommandArguments
): ApplicationCommandOptionData[] | undefined {
  if (!args) return undefined

  return Object.values(args).map((arg) => {
    const base = {
      name: arg.name,
      nameLocalizations: arg.nameLocalizations,
      description: arg.description,
      descriptionLocalizations: arg.descriptionLocalizations,
      required: arg.required ?? false
    }

    if (
      arg.type === ApplicationCommandOptionType.Subcommand ||
      arg.type === ApplicationCommandOptionType.SubcommandGroup
    ) {
      return Object.assign(base, {
        type: arg.type,
        options: arg.args ? parseArguments(arg.args) : undefined
      }) as ApplicationCommandOptionData
    }

    if (arg.type === ApplicationCommandOptionType.String) {
      return Object.assign(base, {
        type: arg.type,
        choices: arg.choices as
          | ApplicationCommandOptionChoiceData<string>[]
          | undefined,
        minLength: arg.min_length,
        maxLength: arg.max_length,
        autocomplete: arg.autocomplete
      }) as ApplicationCommandOptionData
    }

    if (
      arg.type === ApplicationCommandOptionType.Integer ||
      arg.type === ApplicationCommandOptionType.Number
    ) {
      return Object.assign(base, {
        type: arg.type,
        choices: arg.choices as
          | ApplicationCommandOptionChoiceData<number>[]
          | undefined,
        minValue: arg.min_value,
        maxValue: arg.max_value,
        autocomplete: arg.autocomplete
      }) as ApplicationCommandOptionData
    }

    return Object.assign(base, {
      type: arg.type
    }) as ApplicationCommandOptionData
  })
}

export function createCommand<T extends CommandArguments>(command: Command<T>) {
  return command
}
