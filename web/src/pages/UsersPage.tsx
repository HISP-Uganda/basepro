import React from 'react'
import { Box, Typography } from '@mui/material'
import type { GridColDef } from '@mui/x-data-grid'
import { AppDataGrid, type AppDataGridFetchParams } from '../components/datagrid/AppDataGrid'
import { apiRequest } from '../lib/api'
import { buildListQuery, type PaginatedResponse } from '../lib/pagination'

interface UserRow {
  id: number
  username: string
  isActive: boolean
  roles: string[]
  createdAt: string
  updatedAt: string
}

const COLUMNS: GridColDef<UserRow>[] = [
  { field: 'id', headerName: 'ID', width: 90 },
  { field: 'username', headerName: 'Username', flex: 1, minWidth: 180 },
  {
    field: 'roles',
    headerName: 'Roles',
    flex: 1,
    minWidth: 180,
    sortable: false,
    valueGetter: (_value, row) => (row.roles || []).join(', '),
  },
  {
    field: 'isActive',
    headerName: 'Active',
    width: 120,
    type: 'boolean',
  },
  {
    field: 'createdAt',
    headerName: 'Created',
    minWidth: 190,
    valueGetter: (_value, row) => new Date(row.createdAt).toLocaleString(),
  },
]

export function UsersPage() {
  const fetchUsers = React.useCallback(async (params: AppDataGridFetchParams) => {
    const query = buildListQuery(params)
    const response = await apiRequest<PaginatedResponse<UserRow>>(`/users?${query}`)
    return {
      rows: response.items,
      total: response.totalCount,
    }
  }, [])

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <Box>
        <Typography variant="h5" component="h1" gutterBottom>
          Users
        </Typography>
        <Typography color="text.secondary">Server-side pagination, sorting, and filtering for users.</Typography>
      </Box>
      <Box sx={{ height: 620, width: '100%' }}>
        <AppDataGrid columns={COLUMNS} fetchData={fetchUsers} storageKey="users-table" />
      </Box>
    </Box>
  )
}
