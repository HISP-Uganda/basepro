import React from 'react'
import { Alert, AlertTitle, type AlertColor, Snackbar } from '@mui/material'
import type { AppNotification, AppNotificationOptions } from '../notifications/types'

type SnackbarMessage = {
  message: string
  severity?: AlertColor
  autoHideDuration?: number | null
  title?: string
}

interface SnackbarContextValue {
  showSnackbar: (message: SnackbarMessage) => void
  showAppNotification: (notification: AppNotification) => void
}

const SnackbarContext = React.createContext<SnackbarContextValue | undefined>(undefined)

export function SnackbarProvider({ children }: React.PropsWithChildren) {
  const [open, setOpen] = React.useState(false)
  const [payload, setPayload] = React.useState<SnackbarMessage>({
    message: '',
    severity: 'info',
    autoHideDuration: 4000,
  })

  const showSnackbar = React.useCallback((message: SnackbarMessage) => {
    setPayload({
      severity: 'info',
      autoHideDuration: 4000,
      ...message,
    })
    setOpen(true)
  }, [])

  const showAppNotification = React.useCallback((notification: AppNotification) => {
    const message = notification.requestId
      ? `${notification.message} Request ID: ${notification.requestId}`
      : notification.message
    setPayload({
      message,
      severity: notification.kind,
      autoHideDuration: notification.persistent ? null : (notification.autoHideDuration ?? 4000),
      title: notification.title,
    })
    setOpen(true)
  }, [])

  const handleClose = React.useCallback(() => {
    setOpen(false)
  }, [])

  return (
    <SnackbarContext.Provider value={{ showSnackbar, showAppNotification }}>
      {children}
      <Snackbar open={open} autoHideDuration={payload.autoHideDuration} onClose={handleClose}>
        <Alert onClose={handleClose} severity={payload.severity ?? 'info'} sx={{ width: '100%' }}>
          {payload.title ? <AlertTitle>{payload.title}</AlertTitle> : null}
          {payload.message}
        </Alert>
      </Snackbar>
    </SnackbarContext.Provider>
  )
}

export function useSnackbar() {
  const context = React.useContext(SnackbarContext)
  if (!context) {
    throw new Error('useSnackbar must be used inside SnackbarProvider')
  }

  return context
}

export function useNotify() {
  const { showAppNotification } = useSnackbar()

  const send = React.useCallback(
    (kind: AppNotification['kind'], message: string, options?: AppNotificationOptions) => {
      showAppNotification({
        kind,
        message,
        ...options,
      })
    },
    [showAppNotification],
  )

  return React.useMemo(
    () => ({
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
    }),
    [send],
  )
}
