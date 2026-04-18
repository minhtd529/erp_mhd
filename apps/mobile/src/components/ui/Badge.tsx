import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, radius } from '@/lib/theme';

type Variant = 'default' | 'success' | 'danger' | 'warning' | 'secondary' | 'ghost';

interface BadgeProps {
  label: string;
  variant?: Variant;
}

const VARIANTS: Record<Variant, { bg: string; text: string }> = {
  default:   { bg: '#E8EEF8', text: colors.primary },
  success:   { bg: colors.successLight, text: colors.success },
  danger:    { bg: colors.dangerLight, text: colors.danger },
  warning:   { bg: colors.warningLight, text: colors.warning },
  secondary: { bg: '#FBF4EC', text: '#8B5E2A' },
  ghost:     { bg: '#F0F0F0', text: colors.textSecondary },
};

export function Badge({ label, variant = 'default' }: BadgeProps) {
  const v = VARIANTS[variant];
  return (
    <View style={[styles.base, { backgroundColor: v.bg }]}>
      <Text style={[styles.label, { color: v.text }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  base: {
    borderRadius: radius.full,
    paddingHorizontal: 10,
    paddingVertical: 3,
    alignSelf: 'flex-start',
  },
  label: { fontSize: 11, fontWeight: '600' },
});
