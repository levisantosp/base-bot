import { ApplicationCommandOptionType } from 'discord.js'
import { createCommand } from '@/command/create-command'

export default createCommand({
  name: 'sub',
  description: 'Command with a subcommand',
  args: {
    subcommand: {
      type: ApplicationCommandOptionType.Subcommand,
      name: 'subcommand',
      description: 'Subcommand with options',
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
      }
    }
  },
  async run({ ctx }) {
    const args = ctx.args.subcommand
    if (!args) return

    const opt2 = args.opt2 ?? '(not provided)'
    await ctx.reply(`opt1: ${args.opt1}\nopt2: ${opt2}`)
  }
})
