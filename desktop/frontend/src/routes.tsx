import React from 'react'
import { Card, CardContent, Container, Typography } from '@mui/material'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router'
import App from './App'

function HomeRoute() {
  return (
    <Container
      maxWidth="sm"
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Card sx={{ width: '100%' }}>
        <CardContent>
          <Typography variant="h5" component="h1" gutterBottom>
            Skeleton App Ready
          </Typography>
          <Typography color="text.secondary">
            Wails + React + MUI + TanStack Router
          </Typography>
        </CardContent>
      </Card>
    </Container>
  )
}

function NotFoundRoute() {
  return (
    <Container sx={{ py: 8, textAlign: 'center' }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Not Found
      </Typography>
      <Typography color="text.secondary">The page you requested does not exist.</Typography>
    </Container>
  )
}

const rootRoute = createRootRoute({
  component: App,
  notFoundComponent: NotFoundRoute,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomeRoute,
})

const routeTree = rootRoute.addChildren([indexRoute])

export function createAppRouter(initialEntries: string[] = ['/']) {
  return createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries }),
  })
}

export const router = createAppRouter()

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
