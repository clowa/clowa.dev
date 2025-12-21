type LogLevel = 'INFO' | 'WARN' | 'ERROR' | 'DEBUG'

// interface LogEntry {
//   level: LogLevel
//   timestamp: string
//   message: string
//   details?: unknown
// }

class Logger {
  private isDevelopment = typeof process !== 'undefined' && process.env.NODE_ENV === 'development'

  private formatLog(level: LogLevel, message: string, details?: unknown): string {
    const timestamp = new Date().toISOString()
    const prefix = `[${timestamp}] [${level}]`

    if (details) {
      return `${prefix} ${message} ${JSON.stringify(details)}`
    }
    return `${prefix} ${message}`
  }

  info(message: string, details?: unknown): void {
    const formattedMessage = this.formatLog('INFO', message, details)
    console.log(formattedMessage)
  }

  warn(message: string, details?: unknown): void {
    const formattedMessage = this.formatLog('WARN', message, details)
    console.warn(formattedMessage)
  }

  error(message: string, details?: unknown): void {
    const formattedMessage = this.formatLog('ERROR', message, details)
    console.error(formattedMessage)
  }

  debug(message: string, details?: unknown): void {
    if (this.isDevelopment) {
      const formattedMessage = this.formatLog('DEBUG', message, details)
      console.debug(formattedMessage)
    }
  }

  request(method: string, url: string): void {
    this.info(`${method} request`, { url })
  }

  response(status: number, statusText: string, url: string): void {
    this.info(`Response received`, { status, statusText, url })
  }

  fetchError(error: Error, url: string): void {
    this.error(`Failed to fetch from ${url}`, { message: error.message, name: error.name })
  }
}

export const logger = new Logger()
