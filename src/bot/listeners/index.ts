import clientReady from './client-ready'
import interactionCreate from './interaction-create'
import messageCreate from './message-create'

export const listeners = [clientReady, messageCreate, interactionCreate]
