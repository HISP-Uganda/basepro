import React from 'react'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { clearAuthSnapshot, setAuthSnapshot } from '../auth/state'
import { createAppRouter } from '../routes'
import { SnackbarProvider } from '../ui/snackbar'

vi.mock('@mui/x-data-grid', () => ({
  DataGrid: (props: Record<string, any>) => (
    <div>
      {props.rows.map((row: Record<string, any>) => (
        <div key={String(row.id)}>{row.username ?? row.action ?? row.id}</div>
      ))}
    </div>
  ),
}))

function renderRoute(path: string) {
  const router = createAppRouter([path])
  const queryClient = new QueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <SnackbarProvider>
        <RouterProvider router={router} />
      </SnackbarProvider>
    </QueryClientProvider>,
  )
}

function authenticate() {
  setAuthSnapshot({
    isAuthenticated: true,
    accessToken: 'access-token',
    refreshToken: 'refresh-token',
    user: {
      id: 1,
      username: 'admin',
      roles: ['Admin'],
      permissions: ['users.read', 'audit.read', 'settings.read'],
    },
  })
}

describe('users and audit pages', () => {
  beforeEach(() => {
    window.localStorage.clear()
    clearAuthSnapshot()
    vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8080/api/v1')
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    cleanup()
    window.localStorage.clear()
    clearAuthSnapshot()
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
  })

  it('/users renders mocked API rows', async () => {
    authenticate()
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          items: [{ id: 10, username: 'alice', isActive: true, roles: ['Admin'], createdAt: new Date().toISOString() }],
          totalCount: 1,
          page: 1,
          pageSize: 25,
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    renderRoute('/users')

    expect(await screen.findByRole('heading', { name: 'Users', level: 1 })).toBeInTheDocument()
    expect(await screen.findByText('alice')).toBeInTheDocument()
    await waitFor(() => expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/users?'), expect.anything()))
  })

  it('/audit renders mocked API rows', async () => {
    authenticate()
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          items: [{ id: 20, timestamp: new Date().toISOString(), action: 'auth.login.success' }],
          totalCount: 1,
          page: 1,
          pageSize: 25,
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    )

    renderRoute('/audit')

    expect(await screen.findByRole('heading', { name: 'Audit', level: 1 })).toBeInTheDocument()
    expect(await screen.findByText('auth.login.success')).toBeInTheDocument()
    await waitFor(() => expect(fetch).toHaveBeenCalledWith(expect.stringContaining('/audit?'), expect.anything()))
  })
})
