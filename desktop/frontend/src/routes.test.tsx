import React from 'react'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { describe, expect, it } from 'vitest'
import { createAppRouter } from './routes'

function renderWithRouter(initialPath: string) {
  const router = createAppRouter([initialPath])
  const queryClient = new QueryClient()

  return render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe('app routes', () => {
  it('renders home route', async () => {
    renderWithRouter('/')
    expect(await screen.findByRole('heading', { name: 'Skeleton App Ready' })).toBeInTheDocument()
  })

  it('renders not found for unknown route', async () => {
    renderWithRouter('/missing-route')
    expect(await screen.findByRole('heading', { name: 'Not Found' })).toBeInTheDocument()
  })
})
