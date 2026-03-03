import React from 'react'
import { Alert, Box, Button, Card, CardContent, Stack, TextField, Typography } from '@mui/material'
import { useNavigate } from '@tanstack/react-router'
import { isApiError, useAuth } from '../auth/AuthProvider'
import { apiBaseUrl, appName } from '../lib/env'

export function LoginPage() {
  const navigate = useNavigate()
  const auth = useAuth()
  const [username, setUsername] = React.useState('')
  const [password, setPassword] = React.useState('')
  const [submitting, setSubmitting] = React.useState(false)
  const [errorMessage, setErrorMessage] = React.useState('')

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmitting(true)
    setErrorMessage('')

    try {
      await auth.login(username, password)
      await navigate({ to: '/dashboard', replace: true })
    } catch (error) {
      if (isApiError(error)) {
        const requestId = error.requestId ? ` Request ID: ${error.requestId}` : ''
        setErrorMessage(`${error.message}${requestId}`)
      } else {
        setErrorMessage('Unable to sign in right now. Please try again.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Box sx={{ minHeight: '100vh', display: 'grid', placeItems: 'center', px: 2 }}>
      <Card sx={{ width: '100%', maxWidth: 420 }}>
        <CardContent>
          <Stack spacing={2} component="form" onSubmit={handleSubmit}>
            <Typography variant="h5" component="h1">
              {appName}
            </Typography>
            {!apiBaseUrl && <Alert severity="warning">VITE_API_BASE_URL is not configured.</Alert>}
            <TextField
              label="Username"
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              required
            />
            <TextField
              label="Password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
            {errorMessage ? <Alert severity="error">{errorMessage}</Alert> : null}
            <Button type="submit" variant="contained" disabled={submitting}>
              {submitting ? 'Signing in...' : 'Login'}
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  )
}
