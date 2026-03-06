import React from 'react'
import { Alert, Button, Container, Stack, Typography } from '@mui/material'

export function RouteErrorPage() {
  return (
    <Container sx={{ py: 8 }}>
      <Stack spacing={2}>
        <Typography variant="h4" component="h1">
          Something went wrong
        </Typography>
        <Alert severity="error">We could not load this route right now.</Alert>
        <Button variant="contained" onClick={() => window.location.reload()} sx={{ width: 'fit-content' }}>
          Reload Page
        </Button>
      </Stack>
    </Container>
  )
}
