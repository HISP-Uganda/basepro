import React from 'react'
import {
  AppBar,
  Box,
  Button,
  Drawer,
  List,
  ListItemButton,
  ListItemText,
  Toolbar,
  Typography,
} from '@mui/material'
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router'
import { useAuth } from '../auth/AuthProvider'
import { hasPermission, hasRole } from '../rbac/permissions'

const drawerWidth = 240

function sectionTitle(pathname: string) {
  if (pathname.startsWith('/employees')) {
    return 'Employees'
  }
  if (pathname.startsWith('/leave')) {
    return 'Leave'
  }
  if (pathname.startsWith('/payroll')) {
    return 'Payroll'
  }
  if (pathname.startsWith('/users')) {
    return 'Users'
  }
  if (pathname.startsWith('/settings')) {
    return 'Settings'
  }
  if (pathname.startsWith('/dashboard')) {
    return 'Dashboard'
  }
  return 'BasePro'
}

interface NavItem {
  label: string
  path: string
  visible: boolean
}

export function AppShell() {
  const navigate = useNavigate()
  const { logout } = useAuth()
  const pathname = useRouterState({ select: (state) => state.location.pathname })

  const navItems: NavItem[] = [
    { label: 'Dashboard', path: '/dashboard', visible: true },
    { label: 'Employees', path: '/employees', visible: hasRole('Admin') || hasRole('Manager') },
    { label: 'Leave', path: '/leave', visible: hasRole('Admin') || hasRole('Manager') || hasRole('Staff') },
    { label: 'Payroll', path: '/payroll', visible: hasRole('Admin') || hasRole('Manager') },
    {
      label: 'Users',
      path: '/users',
      visible: hasPermission('users.read') || hasPermission('users.write'),
    },
    {
      label: 'Settings',
      path: '/settings',
      visible: hasPermission('settings.read') || hasPermission('settings.write'),
    },
  ]

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            {sectionTitle(pathname)}
          </Typography>
          <Button color="inherit" onClick={() => void logout()}>
            Logout
          </Button>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar>
          <Typography variant="h6" component="div">
            BasePro
          </Typography>
        </Toolbar>
        <List>
          {navItems
            .filter((item) => item.visible)
            .map((item) => (
              <ListItemButton
                key={item.path}
                selected={pathname.startsWith(item.path)}
                onClick={() => void navigate({ to: item.path })}
              >
                <ListItemText primary={item.label} />
              </ListItemButton>
            ))}
        </List>
      </Drawer>

      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  )
}
