import { describe, expect, it } from 'vitest'
import { normalizeError } from './normalizeError'

describe('normalizeError', () => {
  it('maps backend validation errors to normalized structure', () => {
    const normalized = normalizeError({
      status: 422,
      code: 'VALIDATION_ERROR',
      message: 'Validation failed',
      details: {
        username: ['required'],
      },
      requestId: 'req-422',
    })

    expect(normalized).toEqual({
      type: 'validation',
      message: 'Validation failed',
      fieldErrors: {
        username: ['required'],
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
