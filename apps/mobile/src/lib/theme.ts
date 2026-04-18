export const colors = {
  primary: '#1F3A70',
  primaryLight: '#2E5090',
  secondary: '#D4A574',
  background: '#F5F5F5',
  surface: '#FFFFFF',
  border: '#E0E0E0',
  textPrimary: '#1A1A1A',
  textSecondary: '#5A5A5A',
  success: '#2D6A4F',
  successLight: '#E8F5EF',
  danger: '#9B2226',
  dangerLight: '#FDECEA',
  warning: '#B45309',
  warningLight: '#FEF3C7',
  disabled: '#A0A0A0',
  overlay: 'rgba(0,0,0,0.4)',
};

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
};

export const radius = {
  sm: 4,
  md: 6,
  lg: 8,
  xl: 12,
  full: 999,
};

export const typography = {
  h1: { fontSize: 22, fontWeight: '700' as const, color: colors.textPrimary },
  h2: { fontSize: 18, fontWeight: '700' as const, color: colors.textPrimary },
  h3: { fontSize: 16, fontWeight: '600' as const, color: colors.textPrimary },
  body: { fontSize: 14, fontWeight: '400' as const, color: colors.textPrimary },
  small: { fontSize: 12, fontWeight: '400' as const, color: colors.textSecondary },
  label: { fontSize: 13, fontWeight: '500' as const, color: colors.textPrimary },
  mono: { fontSize: 12, fontFamily: 'monospace', color: colors.textSecondary },
};
