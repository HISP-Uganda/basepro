import React from 'react'
import { Alert, type AlertColor, Snackbar } from '@mui/material'

type SnackbarMessage = {
  message: string
  severity?: AlertColor
  autoHideDuration?: number
}

interface SnackbarContextValue {
  showSnackbar: (message: SnackbarMessage) => void
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

  const handleClose = React.useCallback(() => {
    setOpen(false)
  }, [])

  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}
      <Snackbar open={open} autoHideDuration={payload.autoHideDuration} onClose={handleClose}>
        <Alert onClose={handleClose} severity={payload.severity ?? 'info'} sx={{ width: '100%' }}>
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
