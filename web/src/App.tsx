import React from 'react'
import { Outlet } from '@tanstack/react-router'
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material'
import { AuthProvider } from './auth/AuthProvider'

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#0f4c81',
    },
  },
})

export default function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    </ThemeProvider>
  )
}
