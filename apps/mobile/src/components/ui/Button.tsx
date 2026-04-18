import React from 'react';
import { TouchableOpacity, Text, ActivityIndicator, StyleSheet, ViewStyle, TextStyle } from 'react-native';
import { colors, radius, spacing } from '@/lib/theme';

type Variant = 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps {
  title: string;
  onPress?: () => void;
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  disabled?: boolean;
  style?: ViewStyle;
  textStyle?: TextStyle;
}

export function Button({ title, onPress, variant = 'primary', size = 'md', loading, disabled, style, textStyle }: ButtonProps) {
  const isDisabled = disabled || loading;

  const containerStyle: ViewStyle[] = [
    styles.base,
    styles[`size_${size}`],
    styles[`variant_${variant}`],
    isDisabled && styles.disabled,
    style as ViewStyle,
  ].filter(Boolean) as ViewStyle[];

  const labelStyle: TextStyle[] = [
    styles.label,
    styles[`label_${size}`],
    styles[`label_${variant}`],
    isDisabled && styles.labelDisabled,
    textStyle as TextStyle,
  ].filter(Boolean) as TextStyle[];

  return (
    <TouchableOpacity
      style={containerStyle}
      onPress={onPress}
      disabled={isDisabled}
      activeOpacity={0.75}
    >
      {loading
        ? <ActivityIndicator size="small" color={variant === 'primary' ? '#fff' : colors.primary} />
        : <Text style={labelStyle}>{title}</Text>
      }
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  base: {
    borderRadius: radius.md,
    alignItems: 'center',
    justifyContent: 'center',
    flexDirection: 'row',
    gap: spacing.xs,
  },
  size_sm: { height: 34, paddingHorizontal: spacing.md },
  size_md: { height: 44, paddingHorizontal: spacing.xl },
  size_lg: { height: 52, paddingHorizontal: spacing.xxl },
  variant_primary: { backgroundColor: colors.primary },
  variant_secondary: { backgroundColor: 'transparent', borderWidth: 2, borderColor: colors.primary },
  variant_danger: { backgroundColor: colors.danger },
  variant_ghost: { backgroundColor: 'transparent' },
  variant_outline: { backgroundColor: 'transparent', borderWidth: 1, borderColor: colors.border },
  disabled: { opacity: 0.5 },
  label: { fontWeight: '600' },
  label_sm: { fontSize: 13 },
  label_md: { fontSize: 15 },
  label_lg: { fontSize: 16 },
  label_primary: { color: '#fff' },
  label_secondary: { color: colors.primary },
  label_danger: { color: '#fff' },
  label_ghost: { color: colors.primaryLight },
  label_outline: { color: colors.textPrimary },
  labelDisabled: { opacity: 0.7 },
});
