import React from 'react'
import {
  Box,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import type { GridColDef } from '@mui/x-data-grid'
import { JsonMetadataDialog } from '../components/admin/JsonMetadataDialog'
import { AdminRowActions } from '../components/admin/AdminRowActions'
import { buildAdminListRequestQuery, useAdminListSearch } from '../components/admin/listSearch'
import { AppDataGrid, type AppDataGridFetchParams } from '../components/datagrid/AppDataGrid'
import { apiRequest } from '../lib/api'
import type { PaginatedResponse } from '../lib/pagination'
import { useAppNotify } from '../notifications/facade'

interface AuditRow {
  id: number
  timestamp: string
  actorUserId?: number
  action: string
  entityType?: string
  entityId?: string
  metadata?: unknown
}

const ACTION_FILTER_OPTIONS = [
  '',
  'users.create',
  'users.update',
  'users.reset_password',
  'users.set_active',
  'auth.login.success',
  'auth.login.failure',
]

function compactMetadata(metadata: unknown) {
  if (metadata == null) {
    return 'No metadata'
  }
  if (typeof metadata === 'string') {
    return metadata.length > 72 ? `${metadata.slice(0, 72)}...` : metadata
  }
  try {
    const value = JSON.stringify(metadata)
    return value.length > 72 ? `${value.slice(0, 72)}...` : value
  } catch {
    return String(metadata)
  }
}

export function AuditPage() {
  const notify = useAppNotify()
  const { searchInput, setSearchInput, search } = useAdminListSearch()

  const [action, setAction] = React.useState('')
  const { searchInput: actorUserIdInput, setSearchInput: setActorUserIdInput, search: actorUserId } = useAdminListSearch()
  const { searchInput: dateFromInput, setSearchInput: setDateFromInput, search: dateFrom } = useAdminListSearch()
  const { searchInput: dateToInput, setSearchInput: setDateToInput, search: dateTo } = useAdminListSearch()

  const [metadataDialogOpen, setMetadataDialogOpen] = React.useState(false)
  const [selectedMetadata, setSelectedMetadata] = React.useState<unknown>(null)

  const columns = React.useMemo<GridColDef<AuditRow>[]>(
    () => [
      {
        field: 'timestamp',
        headerName: 'Timestamp',
        width: 210,
        valueGetter: (_value, row) => new Date(row.timestamp).toLocaleString(),
      },
      { field: 'actorUserId', headerName: 'Actor', width: 110 },
      { field: 'action', headerName: 'Action', flex: 1, minWidth: 200 },
      { field: 'entityType', headerName: 'Entity Type', width: 140 },
      { field: 'entityId', headerName: 'Entity ID', width: 120 },
      {
        field: 'metadata',
        headerName: 'Metadata',
        flex: 1,
        minWidth: 260,
        sortable: false,
        valueGetter: (_value, row) => compactMetadata(row.metadata),
      },
      {
        field: 'actions',
        headerName: 'Actions',
        sortable: false,
        filterable: false,
        width: 96,
        renderCell: (params) => (
          <AdminRowActions
            rowLabel={params.row.action}
            actions={[
              {
                id: 'view-metadata',
                label: 'View Metadata',
                icon: 'view',
                onClick: () => {
                  setSelectedMetadata(params.row.metadata)
                  setMetadataDialogOpen(true)
                },
              },
            ]}
          />
        ),
      },
    ],
    [],
  )

  const fetchAudit = React.useCallback(
    async (params: AppDataGridFetchParams) => {
      const query = buildAdminListRequestQuery(params, {
        search,
        extra: {
          action,
          actorUserId,
          dateFrom: dateFrom ? `${dateFrom}T00:00:00Z` : undefined,
          dateTo: dateTo ? `${dateTo}T23:59:59Z` : undefined,
        },
      })

      const response = await apiRequest<PaginatedResponse<AuditRow>>(`/audit?${query}`)
      return {
        rows: response.items,
        total: response.totalCount,
      }
    },
    [action, actorUserId, dateFrom, dateTo, search],
  )

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
      <Box>
        <Typography variant="h5" component="h1" gutterBottom>
          Audit Log
        </Typography>
        <Typography color="text.secondary">View audit events with server-side filtering and pagination.</Typography>
      </Box>

      <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5}>
        <TextField
          label="Search"
          placeholder="Search audit action text"
          value={searchInput}
          onChange={(event) => setSearchInput(event.target.value)}
          sx={{ minWidth: 260 }}
        />

        <FormControl sx={{ minWidth: 220 }}>
          <InputLabel id="audit-action-label">Action</InputLabel>
          <Select labelId="audit-action-label" value={action} label="Action" onChange={(event) => setAction(event.target.value)}>
            <MenuItem value="">
              <em>All actions</em>
            </MenuItem>
            {ACTION_FILTER_OPTIONS.filter(Boolean).map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <TextField
          label="Actor User ID"
          value={actorUserIdInput}
          onChange={(event) => setActorUserIdInput(event.target.value)}
          sx={{ minWidth: 180 }}
        />

        <TextField
          label="Date From"
          type="date"
          value={dateFromInput}
          onChange={(event) => setDateFromInput(event.target.value)}
          InputLabelProps={{ shrink: true }}
        />

        <TextField
          label="Date To"
          type="date"
          value={dateToInput}
          onChange={(event) => setDateToInput(event.target.value)}
          InputLabelProps={{ shrink: true }}
        />
      </Stack>

      <Box sx={{ height: 620, width: '100%', minWidth: 0, overflow: 'hidden' }}>
        <AppDataGrid
          columns={columns}
          fetchData={fetchAudit}
          storageKey="audit-table"
          externalQueryKey={`${search}|${action}|${actorUserId}|${dateFrom}|${dateTo}`}
          stickyRightFields={['actions']}
          enablePinnedColumns
        />
      </Box>
      <JsonMetadataDialog
        open={metadataDialogOpen}
        metadata={selectedMetadata}
        onClose={() => setMetadataDialogOpen(false)}
        onCopied={() => notify.success('Metadata copied.')}
      />
    </Box>
  )
}
