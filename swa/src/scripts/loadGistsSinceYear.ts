import { logger } from './logger'

interface Gist {
  id: any
  description: string
  html_url: string
}

async function loadGistsSinceYear(year: number, month: number = 0, day: number = 1): Promise<Gist[]> {
  const headers = new Headers()
  headers.set("Accept", "application/vnd.github+json")
  headers.set("X-GitHub-Api-Version", "2022-11-28")

  if (process.env.GITHUB_TOKEN) {
    headers.set("Authorization", `Bearer ${process.env.GITHUB_TOKEN}`)
  }

  // Get ISO 8601 formatted date string for the year
  const yearString = new Date(new Date().setFullYear(year, month, day)).toISOString()
  const url = `https://api.github.com/users/clowa/gists?since=${yearString}`

  logger.request('GET', url)

  const response = await fetch(url, {
    headers: headers,
  })

  if (!response.ok) {
    logger.response(response.status, response.statusText, url)
    throw new Error(`Network response was not ok ${response.statusText}`)
  }
  logger.response(response.status, response.statusText, url)

  const gists: Gist[] = await response.json()

  // Log fetched gists
  gists.forEach((gist: Gist) => {
    logger.debug(`Gist: ${gist.description}`, { url: gist.html_url })
  })

  return gists
}

export default loadGistsSinceYear
