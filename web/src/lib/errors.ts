import type { ApiError } from './api'

const FALLBACK_ERROR_MESSAGE = 'Something went wrong. Please try again.'

function isApiError(value: unknown): value is ApiError {
  if (!value || typeof value !== 'object') {
    return false
  }

  const candidate = value as Partial<ApiError>
  return typeof candidate.code === 'string' && typeof candidate.message === 'string'
}

export function toUserFriendlyError(error: unknown): string {
  if (!isApiError(error)) {
    return FALLBACK_ERROR_MESSAGE
  }

  let message: string
  switch (error.code) {
    case 'AUTH_UNAUTHORIZED':
      message = 'You are not authorized. Please sign in and try again.'
      break
    case 'AUTH_EXPIRED':
      message = 'Session expired. Please log in again.'
      break
    case 'AUTH_REFRESH_REUSED':
    case 'AUTH_REFRESH_INVALID':
      message = 'Session expired. Please log in again.'
      break
    case 'RATE_LIMITED':
      message = 'Too many requests. Please wait and try again.'
      break
    default:
      message = error.message || FALLBACK_ERROR_MESSAGE
  }

  if (error.requestId) {
    return `${message} Request ID: ${error.requestId}`
  }

  return message
}
