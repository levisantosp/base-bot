import { createListener } from '@/client/create-listener'
import { CommandManager } from '@/command/CommandManager'

export default createListener({
  name: 'interactionCreate',
  async run(client, interaction) {
    if (interaction.isChatInputCommand()) {
      await new CommandManager().exec(client, interaction)
    }
  }
})
