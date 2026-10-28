import { ChannelType, Message } from 'discord.js'
import { createListener } from '@/client/create-listener'
import { CommandManager } from '@/command/CommandManager'

export default createListener({
  name: 'messageCreate',
  async run(client, message) {
    if (message.author.bot || message.channel.type !== ChannelType.GuildText)
      return

    await new CommandManager().exec(client, message as Message<true>)
  }
})
