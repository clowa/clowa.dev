# Overview

This folder contain the [Astro-based](https://docs.astro.build) frontend for the website.

## Favicons and App Icons

The project includes a variety of favicons and app icons to ensure compatibility across different devices and platforms. These icons are located in the `public/` directory and include:

- `favicon.ico`: The standard favicon for browsers.
- `favicon-16x16.png`: A 16x16 pixel PNG version of the favicon for browsers.
- `favicon-32x32.png`: A 32x32 pixel PNG version of the favicon for browsers.
- `favicon-48x48.png`: A 48x48 pixel PNG version of the favicon for Windows and IE.
- `favicon-64x64.png`: A 64x64 pixel PNG version of the favicon for Windows and IE.
- `favicon-128x128.png`: A 128x128 pixel PNG version of the favicon for Chrome Web Store.
- `favicon-180x180.png`/`apple-touch-icon.png`: A 180x180 pixel PNG version of the favicon for Apple devices.
- `favicon-192x192.png`/`android-chrome-192x192.png`: A 192x192 pixel PNG version of the favicon for Android devices.
- `favicon-512x512.png`/`android-chrome-512x512.png`: A 512x512 pixel PNG version of the favicon for standard Progressive Web Apps.

Mostly you will start with a `favicon.svg` file, and use a tool like [RealFaviconGenerator](https://realfavicongenerator.net/), [Favicon Generator](https://design.dev/tools/icon-favicon-builder/) or [vert.sh](https://vert.sh/convert/) to generate the various required sizes and formats.

To figure out which icons are needed for your specific use case, you can refer to [this comprehensive guide on favicons](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs) and [RealFaviconGenerator - FAQ](https://realfavicongenerator.net/faq).

## Web Analytics

This page is using [Rybbit](https://rybbit.io/) for web analytics. To set it up, you need to sign up for an account at Rybbit and create a new site to get your unique Site ID.

To include your Rybbit Site ID in the Astro project, you should set it as an environment variable `PUBLIC_RYBBIT_SITE_ID` in your deployment platform. For local development, you can create a `.env` file in the `swa/` directory with the following content:

```plaintext
PUBLIC_RYBBIT_SITE_ID=your_site_id_here
```

_This variable is rendered into the astro page at buildtime._

You can find more information about using environment variables in Astro in the [Astro documentation](https://docs.astro.build/en/guides/environment-variables/).

## 🧞 Commands

All commands are run from the `swa` directory, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `npm install`             | Installs dependencies                            |
| `npm run dev`             | Starts local dev server at `localhost:4321`      |
| `npm run build`           | Build your production site to `./dist/`          |
| `npm run preview`         | Preview your build locally, before deploying     |
| `npm run astro ...`       | Run CLI commands like `astro add`, `astro check` |
| `npm run astro -- --help` | Get help using the Astro CLI                     |
