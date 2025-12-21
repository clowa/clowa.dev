import { logger } from './logger'

interface Repository {
  id: number
  name: string
  full_name: string
  html_url: string
  description: string
  stargazers_count: number
  forks_count: number
}

async function loadRepositories(repositoryNames: string[]) {
  const headers = new Headers()
  headers.set("Accept", 'application/vnd.github.v3+json')
  headers.set("X-GitHub-Api-Version", "2022-11-28")
  if (process.env.GITHUB_TOKEN) {
    headers.set('Authorization', `Bearer ${process.env.GITHUB_TOKEN}`)
  }

  const repositories: Repository[] = []

  for (const r of repositoryNames) {
    const url = `https://api.github.com/repos/clowa/${r}`
    logger.request('GET', url)

    const response = await fetch(url, {
      headers: headers,
    })

    if (!response.ok) {
      logger.response(response.status, response.statusText, url)
      throw new Error(`Network response was not ok ${response.statusText}`)
    }
    logger.response(response.status, response.statusText, url)

    try {
      const repo: Repository = await response.json()
      repositories.push(repo)
    } catch (error) {
      logger.error(`Failed to parse repository data for ${r}`, error)
      continue
    }
  }

  return repositories
}

export default loadRepositories
