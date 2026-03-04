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
import { type UiThemeMode } from '../preferences'
import { CloseIcon } from '../icons'
import { palettePresets } from './presets'
import { useUiPreferences } from './UiPreferencesProvider'

interface PalettePresetPickerProps {
  open: boolean
  onClose: () => void
}

function paletteColorMain(presetId: string, mode: 'light' | 'dark', key: 'primary' | 'secondary') {
  const preset = palettePresets.find((item) => item.id === presetId) ?? palettePresets[0]
  const palette = mode === 'dark' ? preset.palettes.dark : preset.palettes.light
  const color = palette[key]
  if (color && typeof color === 'object' && 'main' in color && typeof color.main === 'string') {
    return color.main
  }
  return 'transparent'
}

export function PalettePresetPicker({ open, onClose }: PalettePresetPickerProps) {
  const { prefs, resolvedMode, setMode, setPreset } = useUiPreferences()

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ pr: 6 }}>
        Appearance
        <IconButton aria-label="Close appearance dialog" onClick={onClose} sx={{ position: 'absolute', right: 8, top: 8 }}>
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
              value={prefs.mode}
              onChange={(event) => setMode(event.target.value as UiThemeMode)}
            >
              <MenuItem value="light">Light</MenuItem>
              <MenuItem value="dark">Dark</MenuItem>
              <MenuItem value="system">System</MenuItem>
            </Select>
          </FormControl>

          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              Palette preset
            </Typography>
            <Grid container spacing={1.25}>
              {palettePresets.map((preset) => {
                const selected = preset.id === prefs.preset
                const primary = paletteColorMain(preset.id, resolvedMode, 'primary')
                const secondary = paletteColorMain(preset.id, resolvedMode, 'secondary')

                return (
                  <Grid key={preset.id} size={{ xs: 6, sm: 4 }}>
                    <Tooltip title={preset.name}>
                      <Box
                        onClick={() => setPreset(preset.id)}
                        role="button"
                        tabIndex={0}
                        onKeyDown={(event) => {
                          if (event.key === 'Enter' || event.key === ' ') {
                            event.preventDefault()
                            setPreset(preset.id)
                          }
                        }}
                        aria-label={`Select ${preset.name} preset`}
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
                            bgcolor: primary,
                          }}
                        />
                        <Box
                          sx={{
                            width: 18,
                            height: 18,
                            borderRadius: '50%',
                            bgcolor: secondary,
                          }}
                        />
                        <Typography variant="body2" noWrap>
                          {preset.name}
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
