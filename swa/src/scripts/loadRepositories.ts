import { logger } from './logger'
import { assertGitHubOk } from './github'

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

    assertGitHubOk(response, url)
    logger.response(response.status, response.statusText, url)

    const repo: Repository = await response.json()
    repositories.push(repo)
  }

  return repositories
}

export default loadRepositories
