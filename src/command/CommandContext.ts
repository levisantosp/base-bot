import {
  BaseInteraction,
  Message,
  type ChatInputCommandInteraction,
  type InteractionEditReplyOptions,
  type InteractionReplyOptions,
  type MessageEditOptions,
  type MessageReplyOptions,
  type User,
  type Guild
} from 'discord.js'
import type { Client } from '../client/Client'

type CommandData = ChatInputCommandInteraction | Message<true>

type CommandContextOptions<T> = {
  client: Client
  guild: Guild
  data: CommandData
  args: T
  author: User
}

export class CommandContext<T> {
  client: Client
  guild: Guild
  data: CommandData
  args: T
  author: User

  constructor(options: CommandContextOptions<T>) {
    this.client = options.client
    this.guild = options.guild
    this.data = options.data
    this.args = options.args
    this.author = options.author
  }

  async reply(content: string | InteractionReplyOptions | MessageReplyOptions) {
    const payload = typeof content === 'string' ? { content } : content
    if (this.data instanceof BaseInteraction) {
      if (this.data.replied || this.data.deferred) {
        return this.data.followUp(payload as InteractionReplyOptions)
      }
      return this.data.reply(payload as InteractionReplyOptions)
    }
    return this.data.reply(payload as MessageReplyOptions)
  }

  async edit(
    content: string | InteractionEditReplyOptions | MessageEditOptions
  ) {
    const payload = typeof content === 'string' ? { content } : content
    if (this.data instanceof BaseInteraction) {
      return this.data.editReply(payload as InteractionEditReplyOptions)
    }
    return this.data.edit(payload as MessageEditOptions)
  }
}
