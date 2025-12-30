/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly PUBLIC_RYBBIT_SITE_ID: number
  // more env variables...
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
