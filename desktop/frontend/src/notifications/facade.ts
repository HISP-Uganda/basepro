import { dispatchNotification } from './store'
import type { AppNotificationKind, AppNotificationOptions } from './types'

function send(kind: AppNotificationKind, message: string, options?: AppNotificationOptions) {
  dispatchNotification({
    kind,
    message,
    ...options,
  })
}

export const notify = {
  success(message: string, options?: AppNotificationOptions) {
    send('success', message, options)
  },
  error(message: string, options?: AppNotificationOptions) {
    send('error', message, options)
  },
  warning(message: string, options?: AppNotificationOptions) {
    send('warning', message, options)
  },
  info(message: string, options?: AppNotificationOptions) {
    send('info', message, options)
  },
}
