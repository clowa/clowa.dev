import { defineConfig } from 'astro/config';
import { remarkAlert } from "remark-github-blockquote-alert";

import sitemap from '@astrojs/sitemap';

// https://astro.build/config
export default defineConfig({
  site: 'https://clowa.dev',
  markdown: {
    remarkPlugins: [remarkAlert],
  },
  integrations: [sitemap({
    customPages: [],
  })],
  cacheDir: './.cache'
});
