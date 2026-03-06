import React from 'react'
import { Alert, AlertTitle, Snackbar } from '@mui/material'
import { Outlet, useNavigate, useRouter } from '@tanstack/react-router'
import { onSessionExpired } from './auth/session'
import { rememberIntendedDestination } from './auth/sessionExpiry'
import { notify } from './notifications/facade'
import { subscribeNotifications } from './notifications/store'
import type { AppNotification } from './notifications/types'

function App() {
  const router = useRouter()
  const navigate = useNavigate()
  const [notification, setNotification] = React.useState<AppNotification | null>(null)

  React.useEffect(() => {
    return onSessionExpired((reason) => {
      if (reason === 'expired') {
        const location = router.state.location
        rememberIntendedDestination(`${location.pathname}${location.searchStr}${location.hash}`)
        notify.warning('Session expired. Please log in again.', { persistent: true })
        void navigate({ to: '/login', replace: true })
        return
      }

      notify.error('Unable to reach API. Check your connection.')
    })
  }, [navigate, router])

  React.useEffect(() => subscribeNotifications(setNotification), [])

  const notificationText = notification?.requestId
    ? `${notification.message} Request ID: ${notification.requestId}`
    : notification?.message

  return (
    <>
      <Outlet />
      <Snackbar
        open={Boolean(notification)}
        autoHideDuration={notification?.persistent ? null : (notification?.autoHideDuration ?? 4000)}
        onClose={() => setNotification(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity={notification?.kind ?? 'info'} onClose={() => setNotification(null)}>
          {notification?.title ? <AlertTitle>{notification.title}</AlertTitle> : null}
          {notificationText}
        </Alert>
      </Snackbar>
    </>
  )
}

export default App
