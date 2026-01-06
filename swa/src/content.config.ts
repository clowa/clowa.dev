import { defineCollection } from 'astro:content'
import { glob } from 'astro/loaders'
import { z } from 'astro/zod'

const posts = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/pages/posts" }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    author: z.string(),
    draft: z.boolean().optional().default(false),
    pubDate: z.coerce.date(),
    tags: z.array(z.string()).optional(),
  }),
})

export const collections = { posts }
