import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, spacing } from '@/lib/theme';

interface EmptyStateProps {
  message?: string;
  icon?: string;
}

export function EmptyState({ message = 'Không có dữ liệu', icon = '📭' }: EmptyStateProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.icon}>{icon}</Text>
      <Text style={styles.message}>{message}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: 'center', justifyContent: 'center', paddingVertical: 48, gap: spacing.md },
  icon: { fontSize: 36 },
  message: { fontSize: 14, color: colors.textSecondary },
});
