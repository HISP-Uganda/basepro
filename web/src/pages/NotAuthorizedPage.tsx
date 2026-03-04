import React from 'react'
import { Alert, Paper, Stack, Typography } from '@mui/material'

export function NotAuthorizedPage() {
  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Typography variant="h4" component="h1">
          Not Authorized
        </Typography>
        <Alert severity="warning">You do not have permission to access this page.</Alert>
      </Stack>
    </Paper>
  )
}
