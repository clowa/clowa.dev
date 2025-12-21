import { logger } from './logger'

// Timeout for http call
// Quote API can take up to 30 seconds on a double cold start
const timeout: number = 8000

interface Quote {
  content: string
  author: string
}

interface ApiResponse {
  content: string
  author: string
}

// Default quote in case the API is not available.
let quote: Quote = {
  content: "I did a dangerous thing for a man in my position: I decided to tell the truth. 📡",
  author: "Edward Snowden"
}

// Get current website name
const canonicalURL: string = location.protocol + "//" + location.host
console.log("Detected website name " + canonicalURL)

try {
  const controller = new AbortController()
  const id = setTimeout(() => controller.abort(), timeout)
  const url = canonicalURL + "/api/quote"
  logger.request('GET', url)
  const response = await fetch(url, { signal: controller.signal })
  clearTimeout(id)
  logger.response(response.status, response.statusText, url)

  // Check if the API is available and default to a static quote if not.
  if (response.ok) {
    const data: ApiResponse = await response.json()
    quote = {
      content: data.content,
      author: data.author
    }
  }
} catch (error) {
  if (error instanceof DOMException && error.name === 'AbortError') {
    logger.warn(`Request timed out`, { timeout })
  } else if (error instanceof Error) {
    logger.fetchError(error, canonicalURL + "/api/quote")
  } else {
    logger.error('Failed to fetch quote', error)
  }
}

const quoteTextElement = document.getElementById("quote-text")
const quoteAuthorElement = document.getElementById("quote-author")
const quoteLoaderElement = document.getElementById("quote-loader")
const quoteElement = document.getElementById("quote")

if (quoteTextElement) {
  quoteTextElement.innerText = quote.content
}

if (quoteAuthorElement) {
  quoteAuthorElement.innerText = quote.author
}

if (quoteLoaderElement) {
  quoteLoaderElement.remove()
}

if (quoteElement) {
  quoteElement.style.display = "block"
}

export { }
