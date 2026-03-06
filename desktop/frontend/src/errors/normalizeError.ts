import { ApiError } from '../api/client'
import type { NormalizedAppError, NormalizedAppErrorType } from './types'

function toFieldErrors(details: unknown): Record<string, string[]> | undefined {
  if (!details || typeof details !== 'object') {
    return undefined
  }
  const parsed: Record<string, string[]> = {}
  for (const [key, value] of Object.entries(details as Record<string, unknown>)) {
    if (typeof value === 'string' && value.trim()) {
      parsed[key] = [value]
      continue
    }
    if (Array.isArray(value)) {
      const messages = value.filter((entry): entry is string => typeof entry === 'string' && entry.trim().length > 0)
      if (messages.length) {
        parsed[key] = messages
      }
    }
  }
  return Object.keys(parsed).length ? parsed : undefined
}

function normalizeFromApiError(error: ApiError): NormalizedAppErrorType {
  const code = (error.code ?? '').toUpperCase()
  if (error.status === 422 || code === 'VALIDATION_ERROR') {
    return 'validation'
  }
  if (
    error.status === 401 ||
    code === 'AUTH_UNAUTHORIZED' ||
    code === 'AUTH_EXPIRED' ||
    code === 'AUTH_REFRESH_INVALID' ||
    code === 'AUTH_REFRESH_REUSED'
  ) {
    return 'unauthorized'
  }
  if (error.status === 403 || code === 'AUTH_FORBIDDEN' || code === 'FORBIDDEN') {
    return 'forbidden'
  }
  if (error.status === 404 || code === 'NOT_FOUND') {
    return 'not_found'
  }
  if (error.status === 409 || code === 'CONFLICT') {
    return 'conflict'
  }
  if (error.status >= 500 || code.startsWith('SERVER_')) {
    return 'server'
  }
  return 'unknown'
}

export function normalizeError(error: unknown, fallbackMessage = 'Something went wrong. Please try again.'): NormalizedAppError {
  if (error instanceof ApiError) {
    return {
      type: normalizeFromApiError(error),
      message: error.message || fallbackMessage,
      fieldErrors: toFieldErrors(error.details),
      requestId: error.requestId,
    }
  }

  if (error instanceof DOMException && error.name === 'AbortError') {
    return {
      type: 'timeout',
      message: 'The request timed out. Please try again.',
    }
  }

  if (error instanceof TypeError) {
    return {
      type: 'network',
      message: 'Unable to reach the API. Check your connection and try again.',
    }
  }

  if (error instanceof Error) {
    return {
      type: 'unknown',
      message: error.message || fallbackMessage,
    }
  }

  return {
    type: 'unknown',
    message: fallbackMessage,
  }
}
