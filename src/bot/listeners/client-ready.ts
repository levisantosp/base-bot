import { createListener } from '@/client/create-listener'

export default createListener({
  name: 'clientReady',
  async run(client) {
    console.info(`${client.user?.tag} ready!`)

    if (!client.shard?.ids[0]) {
      await client.postCommands()
    }
  }
})
