import { logger } from './logger'

// assertGitHubOk throws a descriptive error when a GitHub API response is not OK.
// These loaders run at build time, so throwing fails the build on purpose: a page
// that silently dropped its repositories/gists (e.g. because the unauthenticated
// 60 req/h rate limit was hit) must not ship. The message names the likely cause
// so a failed build is diagnosable rather than a generic "not ok".
export function assertGitHubOk(response: Response, url: string): void {
  if (response.ok) {
    return
  }

  logger.response(response.status, response.statusText, url)

  const remaining = response.headers.get('x-ratelimit-remaining')
  if ((response.status === 403 || response.status === 429) && remaining === '0') {
    const reset = response.headers.get('x-ratelimit-reset')
    const resetHint = reset
      ? ` (resets at ${new Date(Number(reset) * 1000).toISOString()})`
      : ''
    throw new Error(
      `GitHub API rate limit exceeded${resetHint} for ${url}. ` +
      `Set GITHUB_TOKEN at build time to raise the unauthenticated 60 req/h limit.`,
    )
  }

  throw new Error(`GitHub API request failed: ${response.status} ${response.statusText} for ${url}`)
}
