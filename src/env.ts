import { z } from 'zod'

export const env = z
  .object({
    BOT_TOKEN: z.string(),
    PREFIX: z.string()
  })
  .parse(Bun.env)
