import type { GridFilterModel, GridSortModel } from '@mui/x-data-grid'

export interface PaginatedResponse<T> {
  items: T[]
  totalCount: number
  page: number
  pageSize: number
}

export interface ListQueryInput {
  page: number
  pageSize: number
  sortModel: GridSortModel
  filterModel: GridFilterModel
}

// Server query contract (web + desktop):
// - page is 1-based
// - sort uses "sort=<field>:<asc|desc>"
// - filter uses first non-empty item as "filter=<field>:<value>"
export function buildListQuery(input: ListQueryInput) {
  const query = new URLSearchParams({
    page: String(input.page),
    pageSize: String(input.pageSize),
  })

  const firstSort = input.sortModel[0]
  if (firstSort?.field && firstSort.sort) {
    query.set('sort', `${firstSort.field}:${firstSort.sort}`)
  }

  const firstFilter = input.filterModel.items.find(
    (item) => item.field && item.value !== undefined && item.value !== null && String(item.value).trim() !== '',
  )
  if (firstFilter?.field && firstFilter.value !== undefined && firstFilter.value !== null) {
    query.set('filter', `${firstFilter.field}:${String(firstFilter.value).trim()}`)
  }

  return query.toString()
}
