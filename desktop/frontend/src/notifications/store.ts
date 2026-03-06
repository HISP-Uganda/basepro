import type { AppNotification } from './types'

const listeners = new Set<(notification: AppNotification) => void>()

export function dispatchNotification(notification: AppNotification) {
  for (const listener of listeners) {
    listener(notification)
  }
}

export function subscribeNotifications(listener: (notification: AppNotification) => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}
