import { describe, expect, it, vi } from 'vitest'
import { ApiError } from '../api/client'
import { handleAppError } from './handleAppError'

describe('handleAppError', () => {
  it('maps validation errors to form callback without user notification', async () => {
    const onValidationError = vi.fn()

    const result = await handleAppError(
      new ApiError(422, 'Validation failed', 'VALIDATION_ERROR', { username: ['required'] }),
      {
        onValidationError,
        notifyValidationError: false,
      },
    )

    expect(result.error.type).toBe('validation')
    expect(onValidationError).toHaveBeenCalledWith({ username: ['required'] })
  })

  it('runs unauthorized session-expiry callback', async () => {
    const onSessionExpired = vi.fn(async () => undefined)
    const result = await handleAppError(new ApiError(401, 'Session expired', 'AUTH_EXPIRED'), {
      onSessionExpired,
      notifyUser: false,
    })

    expect(result.error.type).toBe('unauthorized')
    expect(result.didHandleSessionExpiry).toBe(true)
    expect(onSessionExpired).toHaveBeenCalledTimes(1)
  })
})
