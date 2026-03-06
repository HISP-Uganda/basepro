import React from 'react'
import { buildServerQuery } from '../../api/pagination'
import type { AppDataGridFetchParams } from '../datagrid/AppDataGrid'

interface BuildAdminListRequestOptions {
  search?: string
  extra?: Record<string, string | undefined>
}

export function useDebouncedValue(value: string, delayMs: number) {
  const [debounced, setDebounced] = React.useState(value)

  React.useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebounced(value)
    }, delayMs)
    return () => window.clearTimeout(timer)
  }, [delayMs, value])

  return debounced
}

export function useAdminListSearch(initialValue = '', delayMs = 300) {
  const [searchInput, setSearchInput] = React.useState(initialValue)
  const search = useDebouncedValue(searchInput.trim(), delayMs)

  return {
    searchInput,
    setSearchInput,
    search,
  }
}

export function buildAdminListRequestQuery(params: AppDataGridFetchParams, options?: BuildAdminListRequestOptions) {
  const query = new URLSearchParams(buildServerQuery(params))

  const search = options?.search?.trim()
  if (search) {
    query.set('q', search)
  }

  if (options?.extra) {
    for (const [key, value] of Object.entries(options.extra)) {
      const trimmed = value?.trim()
      if (trimmed) {
        query.set(key, trimmed)
      }
    }
  }

  return query.toString()
}
