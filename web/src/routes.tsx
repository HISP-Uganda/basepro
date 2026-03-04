import React from 'react'
import {
  createBrowserHistory,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  redirect,
  type RouterHistory,
} from '@tanstack/react-router'
import App from './App'
import { getAuthSnapshot } from './auth/state'
import { AppShell } from './components/AppShell'
import { AuditPage } from './pages/AuditPage'
import { DashboardPage } from './pages/DashboardPage'
import { EmployeesPage } from './pages/EmployeesPage'
import { LeavePage } from './pages/LeavePage'
import { LoginPage } from './pages/LoginPage'
import { NotAuthorizedPage } from './pages/NotAuthorizedPage'
import { NotFoundPage } from './pages/NotFoundPage'
import { PayrollPage } from './pages/PayrollPage'
import { RouteErrorPage } from './pages/RouteErrorPage'
import { SettingsPage } from './pages/SettingsPage'
import { UsersPage } from './pages/UsersPage'
import { hasPermission, hasRole } from './rbac/permissions'

const rootRoute = createRootRoute({
  component: App,
  notFoundComponent: NotFoundPage,
  errorComponent: RouteErrorPage,
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  beforeLoad: () => {
    if (getAuthSnapshot().isAuthenticated) {
      throw redirect({ to: '/dashboard' })
    }
    throw redirect({ to: '/login' })
  },
  component: () => null,
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  beforeLoad: () => {
    if (getAuthSnapshot().isAuthenticated) {
      throw redirect({ to: '/dashboard' })
    }
  },
  component: LoginPage,
})

const authenticatedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'authenticated',
  beforeLoad: () => {
    if (!getAuthSnapshot().isAuthenticated) {
      throw redirect({ to: '/login' })
    }
  },
  component: AppShell,
})

const dashboardRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/dashboard',
  component: DashboardPage,
})

const employeesRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/employees',
  component: () => (hasRole('Admin') || hasRole('Manager') ? <EmployeesPage /> : <NotAuthorizedPage />),
})

const leaveRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/leave',
  component: () => (hasRole('Admin') || hasRole('Manager') || hasRole('Staff') ? <LeavePage /> : <NotAuthorizedPage />),
})

const payrollRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/payroll',
  component: () => (hasRole('Admin') || hasRole('Manager') ? <PayrollPage /> : <NotAuthorizedPage />),
})

const usersRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/users',
  component: () => (hasPermission('users.read') || hasPermission('users.write') ? <UsersPage /> : <NotAuthorizedPage />),
})

const auditRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/audit',
  component: () => (hasPermission('audit.read') ? <AuditPage /> : <NotAuthorizedPage />),
})

const settingsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: '/settings',
  component: () => (hasPermission('settings.read') || hasPermission('settings.write') ? <SettingsPage /> : <NotAuthorizedPage />),
})

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  authenticatedRoute.addChildren([
    dashboardRoute,
    employeesRoute,
    leaveRoute,
    payrollRoute,
    usersRoute,
    auditRoute,
    settingsRoute,
  ]),
])

export function createAppRouter(
  initialEntries: string[] = ['/'],
  history: RouterHistory = createMemoryHistory({ initialEntries }),
) {
  return createRouter({
    routeTree,
    history,
  })
}

export const router = createAppRouter(['/'], createBrowserHistory())

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
