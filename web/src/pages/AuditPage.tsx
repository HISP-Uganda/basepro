import React from 'react'
import { Box, Typography } from '@mui/material'
import type { GridColDef } from '@mui/x-data-grid'
import { AppDataGrid, type AppDataGridFetchParams } from '../components/datagrid/AppDataGrid'
import { apiRequest } from '../lib/api'
import { buildListQuery, type PaginatedResponse } from '../lib/pagination'

interface AuditRow {
  id: number
  timestamp: string
  actorUserId?: number
  action: string
  entityType?: string
  entityId?: string
}

const COLUMNS: GridColDef<AuditRow>[] = [
  { field: 'id', headerName: 'ID', width: 90 },
  {
    field: 'timestamp',
    headerName: 'Timestamp',
    minWidth: 220,
    valueGetter: (_value, row) => new Date(row.timestamp).toLocaleString(),
  },
  { field: 'actorUserId', headerName: 'Actor User', width: 120 },
  { field: 'action', headerName: 'Action', minWidth: 220, flex: 1 },
  { field: 'entityType', headerName: 'Entity Type', width: 140 },
  { field: 'entityId', headerName: 'Entity ID', width: 120 },
]

export function AuditPage() {
  const fetchAudit = React.useCallback(async (params: AppDataGridFetchParams) => {
    const query = buildListQuery(params)
    const response = await apiRequest<PaginatedResponse<AuditRow>>(`/audit?${query}`)
    return {
      rows: response.items,
      total: response.totalCount,
    }
  }, [])

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <Box>
        <Typography variant="h5" component="h1" gutterBottom>
          Audit
        </Typography>
        <Typography color="text.secondary">Server-side pagination, sorting, and filtering for audit logs.</Typography>
      </Box>
      <Box sx={{ height: 620, width: '100%' }}>
        <AppDataGrid columns={COLUMNS} fetchData={fetchAudit} storageKey="audit-table" />
      </Box>
    </Box>
  )
}
