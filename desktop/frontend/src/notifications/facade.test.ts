import { describe, expect, it, vi } from 'vitest'
import { notify } from './facade'
import { subscribeNotifications } from './store'

describe('notification facade', () => {
  it('dispatches success notification payload', () => {
    const listener = vi.fn()
    const unsubscribe = subscribeNotifications(listener)

    notify.success('Saved successfully.', {
      title: 'Success',
      requestId: 'req-1',
      autoHideDuration: 2500,
    })

    expect(listener).toHaveBeenCalledWith({
      kind: 'success',
      message: 'Saved successfully.',
      title: 'Success',
      requestId: 'req-1',
      autoHideDuration: 2500,
    })

    unsubscribe()
  })
})
