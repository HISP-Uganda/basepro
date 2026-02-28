export interface PalettePreset {
  id: string
  label: string
  primary: string
  secondary: string
  lightBackground: string
  lightPaper: string
  darkBackground: string
  darkPaper: string
}

export const palettePresets: PalettePreset[] = [
  {
    id: 'ocean',
    label: 'Ocean',
    primary: '#0B7285',
    secondary: '#0CA678',
    lightBackground: '#F3F9FA',
    lightPaper: '#FFFFFF',
    darkBackground: '#0F1722',
    darkPaper: '#141E2C',
  },
  {
    id: 'ember',
    label: 'Ember',
    primary: '#C44536',
    secondary: '#F39C12',
    lightBackground: '#FFF7F3',
    lightPaper: '#FFFFFF',
    darkBackground: '#1C1211',
    darkPaper: '#271916',
  },
  {
    id: 'forest',
    label: 'Forest',
    primary: '#2B8A3E',
    secondary: '#5C940D',
    lightBackground: '#F4FAF4',
    lightPaper: '#FFFFFF',
    darkBackground: '#111A13',
    darkPaper: '#18241B',
  },
  {
    id: 'graphite',
    label: 'Graphite',
    primary: '#364FC7',
    secondary: '#495057',
    lightBackground: '#F5F7FA',
    lightPaper: '#FFFFFF',
    darkBackground: '#11141B',
    darkPaper: '#171C27',
  },
  {
    id: 'ruby',
    label: 'Ruby',
    primary: '#A61E4D',
    secondary: '#D6336C',
    lightBackground: '#FFF4F8',
    lightPaper: '#FFFFFF',
    darkBackground: '#1E1118',
    darkPaper: '#2A1620',
  },
  {
    id: 'sunset',
    label: 'Sunset',
    primary: '#E8590C',
    secondary: '#FAB005',
    lightBackground: '#FFF8F2',
    lightPaper: '#FFFFFF',
    darkBackground: '#1D140F',
    darkPaper: '#281C14',
  },
  {
    id: 'berry',
    label: 'Berry',
    primary: '#7B2CBF',
    secondary: '#9D4EDD',
    lightBackground: '#FAF5FF',
    lightPaper: '#FFFFFF',
    darkBackground: '#190F24',
    darkPaper: '#231534',
  },
  {
    id: 'slate',
    label: 'Slate',
    primary: '#1D4ED8',
    secondary: '#0891B2',
    lightBackground: '#F3F6FB',
    lightPaper: '#FFFFFF',
    darkBackground: '#0E1621',
    darkPaper: '#15202E',
  },
]

export function getPalettePreset(id: string) {
  return palettePresets.find((preset) => preset.id === id) ?? palettePresets[0]
}
