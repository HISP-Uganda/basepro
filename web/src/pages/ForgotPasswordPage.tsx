import React from 'react'
import { Alert, Button, Link, Stack, TextField } from '@mui/material'
import { useNavigate } from '@tanstack/react-router'
import { handleAppError } from '../errors/handleAppError'
import { apiRequest } from '../lib/api'
import { AuthSplitLayout } from './auth/AuthSplitLayout'
import { useAuthBranding } from './auth/useAuthBranding'

const successMessage = 'If the account exists, password reset instructions have been sent.'

export function ForgotPasswordPage() {
  const navigate = useNavigate()
  const branding = useAuthBranding()
  const [identifier, setIdentifier] = React.useState('')
  const [submitting, setSubmitting] = React.useState(false)
  const [success, setSuccess] = React.useState(false)
  const [errorMessage, setErrorMessage] = React.useState('')
  const [identifierError, setIdentifierError] = React.useState('')

  const resetUrl = React.useMemo(() => {
    if (typeof window === 'undefined') {
      return undefined
    }
    return `${window.location.origin}/reset-password`
  }, [])

  const onSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    setErrorMessage('')
    setIdentifierError('')

    try {
      await apiRequest(
        '/auth/forgot-password',
        {
          method: 'POST',
          body: JSON.stringify({ identifier: identifier.trim(), resetUrl }),
        },
        { withAuth: false, retryOnUnauthorized: false },
      )
      setSuccess(true)
    } catch (error) {
      const { error: normalized } = await handleAppError(error, {
        fallbackMessage: 'Unable to request reset right now. Please try again.',
        notifyUser: false,
        onValidationError: (fieldErrors) => {
          const message = fieldErrors.identifier?.[0] ?? fieldErrors.email?.[0] ?? fieldErrors.username?.[0] ?? ''
          setIdentifierError(message)
        },
      })
      const requestId = normalized.requestId ? ` Request ID: ${normalized.requestId}` : ''
      setErrorMessage(`${normalized.message}${requestId}`)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthSplitLayout
      branding={branding}
      panelTitle="Forgot password"
      panelSubtitle="Enter your username or email and we will start the password reset flow."
    >
      <Stack spacing={2.25} component="form" onSubmit={onSubmit}>
        <TextField
          label="Username or Email"
          value={identifier}
          onChange={(event) => setIdentifier(event.target.value)}
          required
          fullWidth
          error={Boolean(identifierError)}
          helperText={identifierError}
          InputProps={{ sx: { minHeight: 56 } }}
        />

        {success ? <Alert severity="success">{successMessage}</Alert> : null}
        {errorMessage ? <Alert severity="error">{errorMessage}</Alert> : null}

        <Button type="submit" variant="contained" disabled={submitting} size="large" sx={{ minHeight: 52 }}>
          {submitting ? 'Submitting...' : 'Send Reset Instructions'}
        </Button>

        <Link component="button" type="button" underline="hover" onClick={() => void navigate({ to: '/login' })}>
          Back to login
        </Link>
      </Stack>
    </AuthSplitLayout>
  )
}
