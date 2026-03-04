import type { GridColumnVisibilityModel, GridDensity } from '@mui/x-data-grid'

const STORAGE_PREFIX = 'app.datagrid'

export interface DataGridPinnedColumns {
  left: string[]
  right: string[]
}

export interface DataGridPreferencesV1 {
  version: 1
  pageSize: number
  columnVisibility: GridColumnVisibilityModel
  columnOrder: string[]
  density: GridDensity
  pinnedColumns: DataGridPinnedColumns
}

export interface DataGridStorageOptions {
  migrate?: (value: unknown) => Partial<DataGridPreferencesV1> | null
}

export const defaultDataGridPreferences: DataGridPreferencesV1 = {
  version: 1,
  pageSize: 25,
  columnVisibility: {},
  columnOrder: [],
  density: 'standard',
  pinnedColumns: { left: [], right: [] },
}

function storageKeyFor(storageKey: string) {
  return `${STORAGE_PREFIX}.${storageKey}.v1`
}

function sanitizeStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter((item): item is string => typeof item === 'string')
}

function sanitizeDensity(value: unknown, fallback: GridDensity): GridDensity {
  return value === 'compact' || value === 'comfortable' || value === 'standard' ? value : fallback
}

function sanitizePreferences(value: unknown, fallback: DataGridPreferencesV1): DataGridPreferencesV1 {
  const parsed = typeof value === 'object' && value !== null ? (value as Record<string, unknown>) : {}

  const pageSize = typeof parsed.pageSize === 'number' && Number.isFinite(parsed.pageSize) ? parsed.pageSize : fallback.pageSize
  const columnVisibility =
    typeof parsed.columnVisibility === 'object' && parsed.columnVisibility !== null
      ? (parsed.columnVisibility as GridColumnVisibilityModel)
      : fallback.columnVisibility

  const pinnedRaw = typeof parsed.pinnedColumns === 'object' && parsed.pinnedColumns !== null
    ? (parsed.pinnedColumns as Record<string, unknown>)
    : {}

  return {
    version: 1,
    pageSize,
    columnVisibility,
    columnOrder: sanitizeStringArray(parsed.columnOrder),
    density: sanitizeDensity(parsed.density, fallback.density),
    pinnedColumns: {
      left: sanitizeStringArray(pinnedRaw.left),
      right: sanitizeStringArray(pinnedRaw.right),
    },
  }
}

export function loadDataGridPreferences(
  storageKey: string,
  fallback: DataGridPreferencesV1 = defaultDataGridPreferences,
  options?: DataGridStorageOptions,
): DataGridPreferencesV1 {
  if (typeof window === 'undefined') {
    return fallback
  }

  const raw = window.localStorage.getItem(storageKeyFor(storageKey))
  if (!raw) {
    return fallback
  }

  try {
    const parsed = JSON.parse(raw) as unknown
    const migrated = options?.migrate?.(parsed)
    if (migrated) {
      return sanitizePreferences(migrated, fallback)
    }
    return sanitizePreferences(parsed, fallback)
  } catch {
    return fallback
  }
}

export function saveDataGridPreferences(storageKey: string, value: DataGridPreferencesV1) {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(storageKeyFor(storageKey), JSON.stringify(value))
}
