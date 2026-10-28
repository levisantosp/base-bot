import { createCommand } from '@/command/create-command'

export default createCommand({
  name: 'ping',
  description: 'Shows the bot latency',
  async run({ ctx }) {
    await ctx.reply(`Pong! \`${ctx.client.ws.ping}ms\``)
  }
})
