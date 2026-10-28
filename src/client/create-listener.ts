import type { ClientEvents } from 'discord.js'
import type { Client } from './Client'

export type Listener<T extends keyof ClientEvents = keyof ClientEvents> = {
  name: T
  run: (
    client: Client,
    ...args: T extends 'clientReady' ? [] : ClientEvents[T]
  ) => Promise<unknown>
}

export function createListener<T extends keyof ClientEvents>(
  listener: Listener<T>
) {
  return listener
}
