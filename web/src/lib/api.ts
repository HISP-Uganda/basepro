export type ApiError = {
  code: string
  message: string
  details?: unknown
  requestId?: string
}

type AccessTokenProvider = () => string | null | undefined

type ApiLogger = (message: string, metadata?: Record<string, unknown>) => void

interface ConfigureApiClientOptions {
  getAccessToken?: AccessTokenProvider
  logger?: ApiLogger
}

interface ApiErrorEnvelope {
  error?: {
    code?: string
    message?: string
    details?: unknown
  }
}

let getAccessToken: AccessTokenProvider = () => undefined
let logger: ApiLogger | undefined

function normalizeBaseUrl(baseUrl: string) {
  return baseUrl.trim().replace(/\/+$/, '')
}

function sanitizeHeaders(headers: Headers) {
  const safeHeaders: Record<string, string> = {}
  headers.forEach((value, key) => {
    if (key.toLowerCase() === 'authorization') {
      safeHeaders[key] = '[REDACTED]'
      return
    }
    safeHeaders[key] = value
  })
  return safeHeaders
}

function isJsonResponse(response: Response) {
  return response.headers.get('content-type')?.toLowerCase().includes('application/json') ?? false
}

async function parseApiError(response: Response): Promise<ApiError> {
  const requestId = response.headers.get('X-Request-Id') ?? response.headers.get('x-request-id') ?? undefined
  let code = `HTTP_${response.status}`
  let message = `Request failed with status ${response.status}`
  let details: unknown

  if (isJsonResponse(response)) {
    try {
      const payload = (await response.json()) as ApiErrorEnvelope
      if (payload.error?.code) {
        code = payload.error.code
      }
      if (payload.error?.message) {
        message = payload.error.message
      }
      if (payload.error && 'details' in payload.error) {
        details = payload.error.details
      }
    } catch {
      // Keep fallback when body is invalid JSON.
    }
  }

  return {
    code,
    message,
    details,
    requestId,
  }
}

export function configureApiClient(options: ConfigureApiClientOptions) {
  getAccessToken = options.getAccessToken ?? (() => undefined)
  logger = options.logger
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const baseUrl = normalizeBaseUrl(import.meta.env.VITE_API_BASE_URL ?? '')
  if (!baseUrl) {
    throw new Error('VITE_API_BASE_URL is not configured')
  }

  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const accessToken = getAccessToken()
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const requestUrl = `${baseUrl}${path}`
  logger?.('api.request', {
    method: init.method ?? 'GET',
    url: requestUrl,
    headers: sanitizeHeaders(headers),
  })

  const response = await fetch(requestUrl, {
    ...init,
    headers,
  })

  if (!response.ok) {
    throw await parseApiError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  if (!isJsonResponse(response)) {
    return (await response.text()) as T
  }

  return (await response.json()) as T
}
