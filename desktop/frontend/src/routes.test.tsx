import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { clearSession, configureSessionStorage, setSession } from './auth/session'
import { createAppRouter } from './routes'
import { AppThemeProvider } from './ui/theme'
import {
  defaultSettings,
  type AppSettings,
  type SaveSettingsPatch,
  type SettingsStore,
} from './settings/types'

function createMockSettingsStore(seed: AppSettings): SettingsStore & {
  loadSettingsMock: ReturnType<typeof vi.fn>
  saveSettingsMock: ReturnType<typeof vi.fn>
  resetSettingsMock: ReturnType<typeof vi.fn>
} {
  let state = {
    ...seed,
    uiPrefs: {
      ...seed.uiPrefs,
    },
  }

  const loadSettingsMock = vi.fn(async () => state)
  const saveSettingsMock = vi.fn(async (patch: SaveSettingsPatch) => {
    const nextAuthMode = patch.authMode ?? state.authMode
    state = {
      ...state,
      ...patch,
      authMode: nextAuthMode,
      apiToken:
        nextAuthMode === 'password'
          ? undefined
          : patch.apiToken !== undefined
            ? patch.apiToken || undefined
            : state.apiToken,
      refreshToken:
        patch.refreshToken !== undefined ? patch.refreshToken || undefined : state.refreshToken,
      uiPrefs: {
        ...state.uiPrefs,
        ...(patch.uiPrefs ?? {}),
      },
    }
    return state
  })
  const resetSettingsMock = vi.fn(async () => {
    state = { ...defaultSettings, uiPrefs: { ...defaultSettings.uiPrefs } }
    return state
  })

  return {
    loadSettings: loadSettingsMock,
    saveSettings: saveSettingsMock,
    resetSettings: resetSettingsMock,
    loadSettingsMock,
    saveSettingsMock,
    resetSettingsMock,
  }
}

function renderWithRouter(initialPath: string, store: SettingsStore) {
  const router = createAppRouter([initialPath], store)
  const queryClient = new QueryClient()

  return render(
    <QueryClientProvider client={queryClient}>
      <AppThemeProvider store={store}>
        <RouterProvider router={router} />
      </AppThemeProvider>
    </QueryClientProvider>,
  )
}

describe('app shell routes', () => {
  beforeEach(async () => {
    vi.restoreAllMocks()
    await clearSession()
  })

  afterEach(async () => {
    cleanup()
    await clearSession()
  })

  it('renders app shell + dashboard content for authenticated /dashboard', async () => {
    const store = createMockSettingsStore({
      ...defaultSettings,
      apiBaseUrl: 'http://127.0.0.1:8080',
      refreshToken: 'refresh-token',
    })

    configureSessionStorage(store)
    await setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      expiresAt: Date.now() + 60_000,
    })

    renderWithRouter('/dashboard', store)

    expect(await screen.findByRole('heading', { name: 'Dashboard', level: 1 })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: 'Open user menu' })).toBeInTheDocument()
    expect(screen.getByText('BasePro Desktop v0.1.0')).toBeInTheDocument()
  })

  it('navigates to /settings when Settings is clicked in navigation', async () => {
    const store = createMockSettingsStore({
      ...defaultSettings,
      apiBaseUrl: 'http://127.0.0.1:8080',
      refreshToken: 'refresh-token',
    })

    configureSessionStorage(store)
    await setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      expiresAt: Date.now() + 60_000,
    })

    renderWithRouter('/dashboard', store)

    fireEvent.click(await screen.findByRole('button', { name: 'Settings' }))

    expect(await screen.findByRole('heading', { name: 'Settings' })).toBeInTheDocument()
    expect(await screen.findByText('Connection')).toBeInTheDocument()
  })

  it('persists theme mode and reapplies it after reload', async () => {
    const store = createMockSettingsStore({
      ...defaultSettings,
      apiBaseUrl: 'http://127.0.0.1:8080',
      refreshToken: 'refresh-token',
    })

    configureSessionStorage(store)
    await setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      expiresAt: Date.now() + 60_000,
    })

    const view = renderWithRouter('/settings', store)

    const themeModeSelect = await screen.findByRole('combobox', { name: 'Theme mode' })
    fireEvent.mouseDown(themeModeSelect)
    fireEvent.click(await screen.findByRole('option', { name: 'Dark' }))

    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-theme-mode')).toBe('dark')
      expect(store.saveSettingsMock).toHaveBeenCalledWith(
        expect.objectContaining({
          uiPrefs: expect.objectContaining({ themeMode: 'dark' }),
        }),
      )
    })

    view.unmount()
    renderWithRouter('/settings', store)

    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-theme-pref')).toBe('dark')
      expect(document.documentElement.getAttribute('data-theme-mode')).toBe('dark')
    })
  })

  it('persists palette preset selection and reapplies it after reload', async () => {
    const store = createMockSettingsStore({
      ...defaultSettings,
      apiBaseUrl: 'http://127.0.0.1:8080',
      refreshToken: 'refresh-token',
    })

    configureSessionStorage(store)
    await setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      expiresAt: Date.now() + 60_000,
    })

    const view = renderWithRouter('/settings', store)

    fireEvent.click(await screen.findByRole('button', { name: 'Browse all presets' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Select Ember preset' }))

    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-palette-preset')).toBe('ember')
      expect(store.saveSettingsMock).toHaveBeenCalledWith(
        expect.objectContaining({
          uiPrefs: expect.objectContaining({ palettePreset: 'ember' }),
        }),
      )
    })

    view.unmount()
    renderWithRouter('/settings', store)

    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-palette-preset')).toBe('ember')
      expect(screen.getByText('Active preset: Ember')).toBeInTheDocument()
    })
  })
})
