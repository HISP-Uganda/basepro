export type AppNotificationKind = 'success' | 'error' | 'warning' | 'info'

export type AppNotification = {
  kind: AppNotificationKind
  message: string
  title?: string
  autoHideDuration?: number
  requestId?: string
  persistent?: boolean
}

export type AppNotificationOptions = Omit<AppNotification, 'kind' | 'message'>
