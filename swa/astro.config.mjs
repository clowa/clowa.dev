import { defineConfig } from 'astro/config';
import { remarkAlert } from "remark-github-blockquote-alert";

// https://astro.build/config
export default defineConfig({
    markdown: {
    remarkPlugins: [remarkAlert],
  },
  site: 'https://clowa.dev',
});
