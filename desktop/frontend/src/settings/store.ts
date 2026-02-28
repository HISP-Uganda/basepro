import {
  AUTH_MODES,
  THEME_MODES,
  defaultSettings,
  defaultUiPrefs,
  type AppSettings,
  type AuthMode,
  type SaveSettingsPatch,
  type SettingsStore,
  type ThemeMode,
  type UiPrefs,
} from './types'
import { LoadSettings, ResetSettings, SaveSettings } from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'

const hasWailsBindings = () =>
  typeof window !== 'undefined' &&
  typeof window.go !== 'undefined' &&
  typeof window.go.main !== 'undefined' &&
  typeof window.go.main.App !== 'undefined'

function isObjectRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isAuthMode(value: unknown): value is AuthMode {
  return typeof value === 'string' && AUTH_MODES.some((mode) => mode === value)
}

function isThemeMode(value: unknown): value is ThemeMode {
  return typeof value === 'string' && THEME_MODES.some((mode) => mode === value)
}

function readString(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function readBoolean(value: unknown, fallback = false): boolean {
  return typeof value === 'boolean' ? value : fallback
}

function readPositiveInteger(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0
    ? Math.floor(value)
    : fallback
}

function normalizeUiPrefs(input: unknown): UiPrefs {
  const record = isObjectRecord(input) ? input : {}
  const themeMode = isThemeMode(record.themeMode) ? record.themeMode : defaultUiPrefs.themeMode
  const palettePreset = readString(record.palettePreset, defaultUiPrefs.palettePreset).trim()

  return {
    themeMode,
    palettePreset: palettePreset || defaultUiPrefs.palettePreset,
    navCollapsed: readBoolean(record.navCollapsed, defaultUiPrefs.navCollapsed),
  }
}

function normalizeSettings(input: unknown): AppSettings {
  const record = isObjectRecord(input) ? input : {}
  const authMode = isAuthMode(record.authMode) ? record.authMode : 'password'
  const apiToken = readString(record.apiToken).trim()
  const refreshToken = readString(record.refreshToken).trim()

  return {
    apiBaseUrl: readString(record.apiBaseUrl).trim(),
    authMode,
    apiToken: authMode === 'api_token' && apiToken ? apiToken : undefined,
    refreshToken: refreshToken || undefined,
    requestTimeoutSeconds: readPositiveInteger(
      record.requestTimeoutSeconds,
      defaultSettings.requestTimeoutSeconds,
    ),
    uiPrefs: normalizeUiPrefs(record.uiPrefs),
  }
}

export const settingsStore: SettingsStore = {
  async loadSettings() {
    if (!hasWailsBindings()) {
      return defaultSettings
    }
    const settings = await LoadSettings()
    return normalizeSettings(settings)
  },
  async saveSettings(patch: SaveSettingsPatch) {
    if (!hasWailsBindings()) {
      return normalizeSettings({
        ...defaultSettings,
        ...patch,
        uiPrefs: { ...defaultSettings.uiPrefs, ...patch.uiPrefs },
      })
    }
    const settings = await SaveSettings(new main.SettingsPatch(patch))
    return normalizeSettings(settings)
  },
  async resetSettings() {
    if (!hasWailsBindings()) {
      return defaultSettings
    }
    const settings = await ResetSettings()
    return normalizeSettings(settings)
  },
}
