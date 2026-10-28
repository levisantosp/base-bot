import {
  Client as DiscordClient,
  REST,
  Routes,
  type ApplicationCommandData,
  type ClientOptions
} from 'discord.js'
import { commands } from '@/bot/commands'
import { listeners } from '@/bot/listeners'
import { env } from '@/env'
import { parseArguments, type Command } from '../command/create-command'
import type { Listener } from './create-listener'

const rest = new REST().setToken(env.BOT_TOKEN)

export class Client extends DiscordClient {
  commands = new Map<string, Command<any> & { id: string }>()
  aliases = new Map<string, string>()

  constructor(options: ClientOptions) {
    super(options)
  }

  async connect() {
    for (const command of commands) {
      if (this.commands.get(command.name)) {
        console.warn(`There is already a command named '${command.name}'`)
      }

      this.commands.set(command.name, {
        ...command,
        id: ''
      })

      if (command.aliases) {
        for (const alias of command.aliases) {
          this.aliases.set(alias, command.name)
        }
      }
    }

    for (const listener of listeners as Listener[]) {
      if (listener.name === 'clientReady') {
        this.once(listener.name, () =>
          listener.run(this).catch((e) => console.error(e))
        )
      } else {
        this.on(listener.name, (...args) =>
          listener.run(this, ...args).catch((e) => console.error(e))
        )
      }
    }

    await super.login(env.BOT_TOKEN)
  }

  async postCommands() {
    const commands: ApplicationCommandData[] = []

    this.commands.forEach((cmd) => {
      commands.push({
        name: cmd.name,
        nameLocalizations: cmd.nameLocalizations,
        description: cmd.description,
        descriptionLocalizations: cmd.descriptionLocalizations,
        options: parseArguments(cmd.args),
        type: 1
      })
    })

    const transformKeys = (obj: unknown): unknown => {
      if (obj === null || typeof obj !== 'object') return obj
      if (Array.isArray(obj)) return obj.map(transformKeys)

      const transformed: Record<string, unknown> = {}
      for (const [key, value] of Object.entries(obj)) {
        const newKey =
          key === 'nameLocalizations'
            ? 'name_localizations'
            : key === 'descriptionLocalizations'
              ? 'description_localizations'
              : key
        transformed[newKey] = transformKeys(value)
      }

      return transformed
    }

    const res = (await rest.put(Routes.applicationCommands(this.user!.id), {
      body: transformKeys(commands)
    })) as {
      name: string
      id: string
    }[]

    for (const command of res) {
      const cmd = this.commands.get(command.name)
      if (!cmd) continue

      this.commands.set(command.name, {
        ...cmd,
        id: command.id
      })
    }
  }
}

export const client = new Client({
  intents: ['Guilds', 'GuildMessages', 'MessageContent'],
  allowedMentions: {
    repliedUser: true,
    parse: ['users', 'roles']
  }
})
