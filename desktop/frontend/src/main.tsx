import React from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import './style.css'
import { AppErrorBoundary } from './components/AppErrorBoundary'
import { router } from './routes'
import { AppThemeProvider } from './ui/theme'

const container = document.getElementById('root')
const root = createRoot(container!)
const queryClient = new QueryClient()

root.render(
  <React.StrictMode>
    <AppErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <AppThemeProvider>
          <RouterProvider router={router} />
        </AppThemeProvider>
      </QueryClientProvider>
    </AppErrorBoundary>
  </React.StrictMode>,
)
