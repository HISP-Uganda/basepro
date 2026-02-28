import React from 'react'
import {
  Box,
  Dialog,
  DialogContent,
  DialogTitle,
  FormControl,
  Grid,
  IconButton,
  MenuItem,
  Select,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material'
import CloseIcon from '@mui/icons-material/Close'
import { THEME_MODES, type ThemeMode } from '../settings/types'
import { useThemePreferences } from './theme'

interface PalettePresetPickerProps {
  open: boolean
  onClose: () => void
}

export function PalettePresetPicker({ open, onClose }: PalettePresetPickerProps) {
  const { prefs, presets, setPalettePreset, setThemeMode } = useThemePreferences()

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ pr: 6 }}>
        Appearance
        <IconButton
          aria-label="Close appearance dialog"
          onClick={onClose}
          sx={{ position: 'absolute', right: 8, top: 8 }}
        >
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <Stack spacing={3}>
          <FormControl size="small">
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              Theme mode
            </Typography>
            <Select
              inputProps={{ 'aria-label': 'Theme mode' }}
              value={prefs.themeMode}
              onChange={(event) => void setThemeMode(event.target.value as ThemeMode)}
            >
              {THEME_MODES.map((mode) => (
                <MenuItem key={mode} value={mode}>
                  {mode[0].toUpperCase() + mode.slice(1)}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              Palette preset
            </Typography>
            <Grid container spacing={1.25}>
              {presets.map((preset) => {
                const selected = preset.id === prefs.palettePreset
                return (
                  <Grid key={preset.id} size={{ xs: 6, sm: 4 }}>
                    <Tooltip title={preset.label}>
                      <Box
                        onClick={() => void setPalettePreset(preset.id)}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter' || event.key === ' ') {
                            event.preventDefault()
                            void setPalettePreset(preset.id)
                          }
                        }}
                        aria-label={`Select ${preset.label} preset`}
                        sx={{
                          p: 1,
                          borderRadius: 2,
                          border: '2px solid',
                          borderColor: selected ? 'primary.main' : 'divider',
                          cursor: 'pointer',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 1,
                        }}
                      >
                        <Box
                          sx={{
                            width: 18,
                            height: 18,
                            borderRadius: '50%',
                            bgcolor: preset.primary,
                          }}
                        />
                        <Box
                          sx={{
                            width: 18,
                            height: 18,
                            borderRadius: '50%',
                            bgcolor: preset.secondary,
                          }}
                        />
                        <Typography variant="body2" noWrap>
                          {preset.label}
                        </Typography>
                      </Box>
                    </Tooltip>
                  </Grid>
                )
              })}
            </Grid>
          </Box>
        </Stack>
      </DialogContent>
    </Dialog>
  )
}
