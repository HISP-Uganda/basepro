import React from 'react'
import {
  DataGrid,
  type GridColDef,
  type GridColumnOrderChangeParams,
  type GridColumnVisibilityModel,
  type GridDensity,
  type GridFilterModel,
  type GridPaginationModel,
  type GridSortModel,
  type GridValidRowModel,
} from '@mui/x-data-grid'
import type { ApiError } from '../../lib/api'
import { useSnackbar } from '../../ui/snackbar'
import {
  defaultDataGridPreferences,
  loadDataGridPreferences,
  saveDataGridPreferences,
  type DataGridPinnedColumns,
} from './storage'

const PAGE_SIZE_OPTIONS = [10, 25, 50, 100]
export interface AppDataGridFetchParams {
  page: number
  pageSize: number
  sortModel: GridSortModel
  filterModel: GridFilterModel
}

export interface AppDataGridFetchResult<R extends GridValidRowModel = GridValidRowModel> {
  rows: R[]
  total: number
}

interface AppDataGridProps<R extends GridValidRowModel = GridValidRowModel> {
  columns: GridColDef<R>[]
  fetchData: (params: AppDataGridFetchParams) => Promise<AppDataGridFetchResult<R>>
  storageKey: string
  getRowId?: (row: R) => string | number
  reloadToken?: number
  enablePinnedColumns?: boolean
}

function moveField(fields: string[], field: string, targetIndex: number | undefined): string[] {
  const sourceIndex = fields.indexOf(field)
  if (sourceIndex < 0) {
    return fields
  }
  const next = [...fields]
  const [moved] = next.splice(sourceIndex, 1)
  const index = typeof targetIndex === 'number' ? Math.max(0, Math.min(targetIndex, next.length)) : next.length
  next.splice(index, 0, moved)
  return next
}

function applyColumnOrder<R extends GridValidRowModel>(columns: GridColDef<R>[], columnOrder: string[]) {
  if (!columnOrder.length) {
    return columns
  }
  const columnByField = new Map(columns.map((column) => [column.field, column] as const))
  const ordered: GridColDef<R>[] = []

  for (const field of columnOrder) {
    const column = columnByField.get(field)
    if (column) {
      ordered.push(column)
      columnByField.delete(field)
    }
  }
  for (const column of columns) {
    if (columnByField.has(column.field)) {
      ordered.push(column)
    }
  }

  return ordered
}

function isApiError(error: unknown): error is ApiError {
  if (!error || typeof error !== 'object') {
    return false
  }
  const candidate = error as Partial<ApiError>
  return typeof candidate.code === 'string' && typeof candidate.message === 'string'
}

function toErrorMessage(error: unknown) {
  if (isApiError(error)) {
    return error.requestId ? `${error.message} Request ID: ${error.requestId}` : error.message
  }
  return 'Unable to load data.'
}

export function AppDataGrid<R extends GridValidRowModel = GridValidRowModel>({
  columns,
  fetchData,
  storageKey,
  getRowId,
  reloadToken,
  enablePinnedColumns = false,
}: AppDataGridProps<R>) {
  const { showSnackbar } = useSnackbar()
  const [rows, setRows] = React.useState<R[]>([])
  const [rowCount, setRowCount] = React.useState(0)
  const [loading, setLoading] = React.useState(false)
  const [hydrated, setHydrated] = React.useState(false)
  const [paginationModel, setPaginationModel] = React.useState<GridPaginationModel>({
    page: 0,
    pageSize: defaultDataGridPreferences.pageSize,
  })
  const [sortModel, setSortModel] = React.useState<GridSortModel>([])
  const [filterModel, setFilterModel] = React.useState<GridFilterModel>({ items: [] })
  const [columnVisibilityModel, setColumnVisibilityModel] = React.useState<GridColumnVisibilityModel>({})
  const [columnOrder, setColumnOrder] = React.useState<string[]>([])
  const [density, setDensity] = React.useState<GridDensity>(defaultDataGridPreferences.density)
  const [pinnedColumns, setPinnedColumns] = React.useState<DataGridPinnedColumns>(defaultDataGridPreferences.pinnedColumns)
  const requestIdRef = React.useRef(0)

  React.useEffect(() => {
    const preferences = loadDataGridPreferences(storageKey)
    setPaginationModel({
      page: 0,
      pageSize: PAGE_SIZE_OPTIONS.includes(preferences.pageSize) ? preferences.pageSize : defaultDataGridPreferences.pageSize,
    })
    setColumnVisibilityModel(preferences.columnVisibility)
    setColumnOrder(preferences.columnOrder)
    setDensity(preferences.density)
    setPinnedColumns(preferences.pinnedColumns)
    setHydrated(true)
  }, [storageKey])

  React.useEffect(() => {
    if (!hydrated) {
      return
    }
    saveDataGridPreferences(storageKey, {
      version: 1,
      pageSize: paginationModel.pageSize,
      columnVisibility: columnVisibilityModel,
      columnOrder,
      density,
      pinnedColumns,
    })
  }, [hydrated, storageKey, paginationModel.pageSize, columnVisibilityModel, columnOrder, density, pinnedColumns])

  React.useEffect(() => {
    if (!hydrated) {
      return
    }
    const requestId = ++requestIdRef.current
    setLoading(true)

    void fetchData({
      page: paginationModel.page + 1,
      pageSize: paginationModel.pageSize,
      sortModel,
      filterModel,
    })
      .then((result) => {
        if (requestId !== requestIdRef.current) {
          return
        }
        setRows(result.rows)
        setRowCount(result.total)
      })
      .catch((error: unknown) => {
        if (requestId !== requestIdRef.current) {
          return
        }
        showSnackbar({ message: toErrorMessage(error), severity: 'error' })
      })
      .finally(() => {
        if (requestId === requestIdRef.current) {
          setLoading(false)
        }
      })
  }, [hydrated, paginationModel, sortModel, filterModel, reloadToken, fetchData, showSnackbar])

  const orderedColumns = React.useMemo(() => applyColumnOrder(columns, columnOrder), [columns, columnOrder])

  return (
    <DataGrid
      columns={orderedColumns}
      rows={rows}
      rowCount={rowCount}
      loading={loading}
      getRowId={getRowId}
      pagination
      paginationMode="server"
      paginationModel={paginationModel}
      onPaginationModelChange={setPaginationModel}
      pageSizeOptions={PAGE_SIZE_OPTIONS}
      sortingMode="server"
      sortModel={sortModel}
      onSortModelChange={setSortModel}
      filterMode="server"
      filterModel={filterModel}
      onFilterModelChange={setFilterModel}
      columnVisibilityModel={columnVisibilityModel}
      onColumnVisibilityModelChange={setColumnVisibilityModel}
      onColumnOrderChange={(params: GridColumnOrderChangeParams) =>
        setColumnOrder((current) => {
          const baseline = current.length ? current : orderedColumns.map((column) => column.field)
          return moveField(baseline, params.column.field, params.targetIndex)
        })
      }
      density={density}
      onDensityChange={setDensity}
      showToolbar
      slotProps={{
        toolbar: {
          csvOptions: {
            fileName: storageKey.replace(/[^a-z0-9_-]/gi, '_'),
          },
          printOptions: {
            disableToolbarButton: true,
          },
        },
      }}
      sx={{
        '& .MuiDataGrid-columnHeaderTitle': {
          fontWeight: 700,
        },
      }}
      disableRowSelectionOnClick
      {...(enablePinnedColumns
        ? {
            pinnedColumns,
            onPinnedColumnsChange: (value: { left?: string[]; right?: string[] }) =>
              setPinnedColumns({
                left: value.left ?? [],
                right: value.right ?? [],
              }),
          }
        : {})}
    />
  )
}
