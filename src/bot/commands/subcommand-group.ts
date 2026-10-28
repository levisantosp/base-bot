import { ApplicationCommandOptionType } from 'discord.js'
import { createCommand } from '@/command/create-command'

export default createCommand({
  name: 'group',
  description: 'Command with a subcommand group',
  args: {
    group: {
      type: ApplicationCommandOptionType.SubcommandGroup,
      name: 'group',
      description: 'Subcommand group',
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
      }
    }
  },
  async run({ ctx }) {
    await ctx.reply(
      `opt1: ${ctx.args.group?.subcommand?.opt1}\nopt2: ${ctx.args.group?.subcommand?.opt2 ?? '(not provided)'}`
    )
  }
})
