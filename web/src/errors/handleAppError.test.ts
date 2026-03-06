import { describe, expect, it, vi } from 'vitest'
import { handleAppError } from './handleAppError'

describe('handleAppError', () => {
  it('maps validation errors to callback and suppresses validation toast by default', async () => {
    const onValidationError = vi.fn()
    const notifier = {
      error: vi.fn(),
    }

    const result = await handleAppError(
      {
        status: 422,
        code: 'VALIDATION_ERROR',
        message: 'Validation failed',
        details: {
          username: ['required'],
        },
      },
      {
        onValidationError,
        notifier,
      },
    )

    expect(result.error.type).toBe('validation')
    expect(onValidationError).toHaveBeenCalledWith({ username: ['required'] })
    expect(notifier.error).not.toHaveBeenCalled()
  })

  it('runs unauthorized session expiry callback and notifies user', async () => {
    const onSessionExpired = vi.fn(async () => undefined)
    const notifier = {
      error: vi.fn(),
    }

    const result = await handleAppError(
      {
        status: 401,
        code: 'AUTH_EXPIRED',
        message: 'Session expired',
      },
      {
        onSessionExpired,
        notifier,
      },
    )

    expect(result.error.type).toBe('unauthorized')
    expect(result.didHandleSessionExpiry).toBe(true)
    expect(onSessionExpired).toHaveBeenCalledTimes(1)
    expect(notifier.error).toHaveBeenCalledWith('Session expired', { requestId: undefined, persistent: true })
  })
})
