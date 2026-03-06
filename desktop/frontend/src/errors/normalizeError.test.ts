import { describe, expect, it } from 'vitest'
import { ApiError } from '../api/client'
import { normalizeError } from './normalizeError'

describe('normalizeError', () => {
  it('maps validation ApiError details to normalized validation error', () => {
    const error = new ApiError(
      422,
      'Validation failed',
      'VALIDATION_ERROR',
      {
        username: ['required'],
        email: 'invalid',
      },
      'req-422',
    )

    expect(normalizeError(error)).toEqual({
      type: 'validation',
      message: 'Validation failed',
      fieldErrors: {
        username: ['required'],
        email: ['invalid'],
      },
      requestId: 'req-422',
    })
  })

  it('maps network and timeout errors', () => {
    expect(normalizeError(new TypeError('Failed to fetch'))).toMatchObject({
      type: 'network',
    })
    expect(normalizeError(new DOMException('timed out', 'AbortError'))).toMatchObject({
      type: 'timeout',
    })
  })
})
