import React from 'react'
import { Button, Paper, Stack, Typography } from '@mui/material'
import { useNavigate } from '@tanstack/react-router'
import { hasPermission, hasRole } from '../rbac/permissions'

export function DashboardPage() {
  const navigate = useNavigate()

  const moduleActions = [
    {
      label: 'Employees',
      path: '/employees',
      enabled: hasRole('Admin') || hasRole('Manager'),
    },
    {
      label: 'Leave',
      path: '/leave',
      enabled: hasRole('Admin') || hasRole('Manager') || hasRole('Staff'),
    },
    {
      label: 'Payroll',
      path: '/payroll',
      enabled: hasRole('Admin') || hasRole('Manager'),
    },
    {
      label: 'Users',
      path: '/users',
      enabled: hasPermission('users.read') || hasPermission('users.write'),
    },
    {
      label: 'Settings',
      path: '/settings',
      enabled: hasPermission('settings.read') || hasPermission('settings.write'),
    },
  ]

  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Dashboard
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Role-based module access is enforced by the backend. UI visibility is supplemental.
      </Typography>
      <Stack direction="row" spacing={1.5} useFlexGap flexWrap="wrap">
        {moduleActions.map((item) => (
          <Button
            key={item.label}
            variant="contained"
            onClick={() => void navigate({ to: item.path })}
            disabled={!item.enabled}
          >
            {item.label}
          </Button>
        ))}
      </Stack>
    </Paper>
  )
}
