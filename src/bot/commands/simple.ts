import { ApplicationCommandOptionType } from 'discord.js'
import { createCommand } from '@/command/create-command'

export default createCommand({
  name: 'simple',
  aliases: ['s'],
  description: 'Command with two options',
  args: {
    opt1: {
      type: ApplicationCommandOptionType.String,
      name: 'opt1',
      description: 'Required option',
      required: true
    },
    opt2: {
      type: ApplicationCommandOptionType.String,
      name: 'opt2',
      description: 'Optional option'
    }
  },
  async run({ ctx }) {
    const opt2 = ctx.args.opt2 ?? '(not provided)'
    await ctx.reply(`opt1: ${ctx.args.opt1}\nopt2: ${opt2}`)
  }
})
