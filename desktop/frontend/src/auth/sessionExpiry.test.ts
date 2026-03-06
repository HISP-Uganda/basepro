import { beforeEach, describe, expect, it } from 'vitest'
import { consumeIntendedDestination, rememberIntendedDestination } from './sessionExpiry'

describe('session expiry redirect helpers', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
  })

  it('stores and consumes intended destination', () => {
    rememberIntendedDestination('/users?page=2')
    expect(consumeIntendedDestination('/dashboard')).toBe('/users?page=2')
    expect(consumeIntendedDestination('/dashboard')).toBe('/dashboard')
  })

  it('does not store auth route destinations', () => {
    rememberIntendedDestination('/login')
    expect(consumeIntendedDestination('/dashboard')).toBe('/dashboard')
  })
})
