import React from 'react'
import { Alert, Button, Container, Stack, Typography } from '@mui/material'

interface AppErrorBoundaryState {
  hasError: boolean
  errorMessage: string
}

export class AppErrorBoundary extends React.Component<React.PropsWithChildren, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    hasError: false,
    errorMessage: '',
  }

  static getDerivedStateFromError(error: unknown): AppErrorBoundaryState {
    const message = error instanceof Error ? error.message : 'Unexpected application error'
    return {
      hasError: true,
      errorMessage: message,
    }
  }

  componentDidCatch() {
    // Intentionally keep production fallback generic and avoid printing stack traces.
  }

  private reload() {
    window.location.reload()
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    return (
      <Container sx={{ py: 8 }}>
        <Stack spacing={2}>
          <Typography variant="h4" component="h1">
            Something went wrong
          </Typography>
          <Alert severity="error">The app encountered an unexpected issue. Reload to continue.</Alert>
          {import.meta.env.DEV && this.state.errorMessage ? (
            <Typography variant="body2" color="text.secondary">
              {this.state.errorMessage}
            </Typography>
          ) : null}
          <Button variant="contained" onClick={() => this.reload()} sx={{ width: 'fit-content' }}>
            Reload App
          </Button>
        </Stack>
      </Container>
    )
  }
}
