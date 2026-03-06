import { notify } from '../notifications/facade'
import { normalizeError } from './normalizeError'
import type { NormalizedAppError } from './types'

interface HandleAppErrorOptions {
  fallbackMessage?: string
  notifyUser?: boolean
  notifyValidationError?: boolean
  onValidationError?: (fieldErrors: Record<string, string[]>) => void
  onSessionExpired?: () => Promise<void> | void
}

interface HandleAppErrorResult {
  error: NormalizedAppError
  didHandleSessionExpiry: boolean
}

export async function handleAppError(error: unknown, options: HandleAppErrorOptions = {}): Promise<HandleAppErrorResult> {
  const normalized = normalizeError(error, options.fallbackMessage)
  const fieldErrors = normalized.fieldErrors

  if (normalized.type === 'validation' && fieldErrors && options.onValidationError) {
    options.onValidationError(fieldErrors)
  }

  let didHandleSessionExpiry = false
  if (normalized.type === 'unauthorized' && options.onSessionExpired) {
    didHandleSessionExpiry = true
    await options.onSessionExpired()
  }

  const shouldNotify =
    options.notifyUser !== false && (options.notifyValidationError || normalized.type !== 'validation')
  if (shouldNotify) {
    notify.error(normalized.message, {
      requestId: normalized.requestId,
      persistent: normalized.type === 'unauthorized',
    })
  }

  return {
    error: normalized,
    didHandleSessionExpiry,
  }
}
