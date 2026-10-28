import {
  ApplicationCommandOptionType,
  ChatInputCommandInteraction,
  Message,
  type Guild,
  type User
} from 'discord.js'
import { env } from '@/env'
import { Client } from '../client/Client'
import { CommandContext } from './CommandContext'
import type { CommandArguments, ResolveArguments } from './create-command'

const nullCatch = () => null

const getters = {
  [ApplicationCommandOptionType.String]: 'getString',
  [ApplicationCommandOptionType.Integer]: 'getInteger',
  [ApplicationCommandOptionType.Number]: 'getNumber',
  [ApplicationCommandOptionType.Boolean]: 'getBoolean',
  [ApplicationCommandOptionType.User]: 'getUser',
  [ApplicationCommandOptionType.Channel]: 'getChannel',
  [ApplicationCommandOptionType.Role]: 'getRole',
  [ApplicationCommandOptionType.Attachment]: 'getAttachment'
} as const

type ParseMessageArgumentOptions = {
  app: Client
  guild: Guild
  type: number
  value: string | undefined
}

export class CommandManager {
  private async parseMessageArguments(props: ParseMessageArgumentOptions) {
    if (!props.value) return undefined

    switch (props.type) {
      case ApplicationCommandOptionType.Integer:
      case ApplicationCommandOptionType.Number: {
        return Number(props.value)
      }
      case ApplicationCommandOptionType.Boolean: {
        return props.value === 'true'
      }
      case ApplicationCommandOptionType.User: {
        const id = props.value.replace(/[<@!>]/g, '')
        return (
          props.app.users.cache.get(id) ??
          (await props.app.users.fetch(id).catch(nullCatch))
        )
      }
      case ApplicationCommandOptionType.Channel: {
        const id = props.value.replace(/[<#>]/g, '')
        return (
          props.guild.channels.cache.get(id) ??
          (await props.guild.channels.fetch(id).catch(nullCatch))
        )
      }
      case ApplicationCommandOptionType.Role: {
        const id = props.value.replace(/[<@&>]/g, '')
        return (
          props.guild.roles.cache.get(id) ??
          (await props.guild.roles.fetch(id).catch(nullCatch))
        )
      }
      default: {
        return props.value
      }
    }
  }

  private async verifyMessageArguments(
    app: Client,
    guild: Guild,
    args: CommandArguments,
    input: string[],
    namespace: Record<string, unknown>,
    message: Message
  ): Promise<boolean> {
    const keys = Object.keys(args)
    const isSubcommandLevel = keys.some(
      (key) =>
        args[key]!.type === ApplicationCommandOptionType.Subcommand ||
        args[key]!.type === ApplicationCommandOptionType.SubcommandGroup
    )

    if (isSubcommandLevel) {
      const subName = input[0]?.toLowerCase()
      if (!subName) {
        await message.reply(`Missing required argument: ${keys.join('/')}`)
        return false
      }

      const matchedKey = keys.find((key) => args[key]!.name === subName)
      if (matchedKey) {
        const arg = args[matchedKey]!
        input.shift()
        namespace[matchedKey] = {}

        if (arg.args) {
          return await this.verifyMessageArguments(
            app,
            guild,
            arg.args,
            input,
            namespace[matchedKey] as Record<string, unknown>,
            message
          )
        }
      }
      return true
    }

    for (let index = 0; index < keys.length; index++) {
      const key = keys[index]!
      const arg = args[key]!

      const value = await this.parseMessageArguments({
        app,
        guild,
        type: arg.type,
        value: input[index]!
      })

      if (value === undefined && arg.required) {
        await message.reply(`Missing required argument: ${arg.name}`)
        return false
      }

      namespace[key] = value
    }
    return true
  }

  async exec(
    client: Client,
    data: ChatInputCommandInteraction | Message<true>
  ) {
    if (!data.guild || !data.guildId) return

    let commandName = ''
    const args: string[] = []
    let user: User
    const isSlash = data instanceof ChatInputCommandInteraction

    if (isSlash) {
      commandName = data.commandName
      user = data.user
    } else {
      if (!data.content.startsWith(env.PREFIX)) return
      const messageArray = data.content
        .slice(env.PREFIX.length)
        .trim()
        .split(/\s+/)
      commandName = messageArray.shift()?.toLowerCase() ?? ''
      args.push(...messageArray)
      user = data.author
    }

    const command =
      client.commands.get(commandName) ??
      client.commands.get(client.aliases.get(commandName) ?? '')
    if (!command) return

    const parsedArgs: Record<string, unknown> = {}
    const ctx = new CommandContext({
      client: client,
      data,
      guild: data.guild,
      args: parsedArgs as ResolveArguments<CommandArguments>,
      author: user
    })

    if (command.permissions) {
      const permissions = (
        isSlash ? data.memberPermissions : data.member?.permissions
      )?.missing(command.permissions)
      if (permissions?.length) {
        return ctx.reply(
          `You need these permissions: ${permissions.join(', ')}`
        )
      }
    }

    if (command.botPermissions) {
      const permissions = data.guild.members.me?.permissions.missing(
        command.botPermissions
      )
      if (permissions?.length) {
        return ctx.reply(
          `The bot needs these permissions: ${permissions.join(', ')}`
        )
      }
    }

    if (command.args) {
      if (isSlash) {
        this.parseSlashArguments(command.args, data, parsedArgs)
      } else {
        const success = await this.verifyMessageArguments(
          client,
          data.guild,
          command.args,
          args,
          parsedArgs,
          data
        )
        if (!success) return
      }
    }

    if (command.ephemeral && isSlash) {
      await data.deferReply({ ephemeral: true }).catch(undefined)
    } else if (command.isThinking) {
      if (isSlash) await data.deferReply().catch(undefined)
      else await data.channel.sendTyping()
    }

    return command.run({
      ctx,
      client,
      id: command.name
    })
  }

  private parseSlashArguments(
    args: CommandArguments,
    interaction: ChatInputCommandInteraction,
    parsedNamespace: Record<string, unknown>
  ) {
    const group = interaction.options.getSubcommandGroup(false)
    const subcommand = interaction.options.getSubcommand(false)
    let currentDefinitions = args
    let currentParsed = parsedNamespace

    if (group) {
      const key = Object.keys(currentDefinitions).find(
        (name) => currentDefinitions[name]!.name === group
      )
      if (key) {
        currentParsed[key] = {}
        currentParsed = currentParsed[key] as Record<string, unknown>
        currentDefinitions = currentDefinitions[key]!.args ?? {}
      }
    }

    if (subcommand) {
      const key = Object.keys(currentDefinitions).find(
        (name) => currentDefinitions[name]!.name === subcommand
      )
      if (key) {
        currentParsed[key] = {}
        currentParsed = currentParsed[key] as Record<string, unknown>
        currentDefinitions = currentDefinitions[key]!.args ?? {}
      }
    }

    for (const [key, arg] of Object.entries(currentDefinitions)) {
      if (
        arg.type === ApplicationCommandOptionType.Subcommand ||
        arg.type === ApplicationCommandOptionType.SubcommandGroup
      )
        continue

      const getter = getters[arg.type as keyof typeof getters]
      const value = getter
        ? (
            interaction.options as unknown as Record<
              string,
              (name: string) => unknown
            >
          )[getter]!(arg.name)
        : interaction.options.get(arg.name)?.value

      currentParsed[key] = value
    }
  }
}
