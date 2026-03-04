import React from 'react'
import { Paper, Typography } from '@mui/material'

interface ModulePlaceholderPageProps {
  title: string
}

export function ModulePlaceholderPage({ title }: ModulePlaceholderPageProps) {
  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {title}
      </Typography>
      <Typography color="text.secondary">Coming soon</Typography>
    </Paper>
  )
}
