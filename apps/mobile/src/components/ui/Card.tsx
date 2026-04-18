import React from 'react';
import { View, Text, ViewStyle, StyleSheet } from 'react-native';
import { colors, radius, spacing } from '@/lib/theme';

interface CardProps {
  children: React.ReactNode;
  style?: ViewStyle;
}

export function Card({ children, style }: CardProps) {
  return <View style={[styles.card, style]}>{children}</View>;
}

interface StatCardProps {
  label: string;
  value: string;
  sub?: string;
  accent?: string;
}

export function StatCard({ label, value, sub, accent }: StatCardProps) {
  return (
    <Card style={styles.statCard}>
      <View style={[styles.statAccent, { backgroundColor: accent ?? colors.primary + '18' }]}>
        <Text style={[styles.statAccentDot, { color: accent ? '#fff' : colors.primary }]}>●</Text>
      </View>
      <Text style={styles.statLabel}>{label}</Text>
      <Text style={styles.statValue}>{value}</Text>
      {sub && <Text style={styles.statSub}>{sub}</Text>}
    </Card>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    padding: spacing.lg,
    borderWidth: 1,
    borderColor: colors.border,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.06,
    shadowRadius: 3,
    elevation: 2,
  },
  statCard: {
    flex: 1,
    gap: spacing.xs,
  },
  statAccent: {
    width: 32,
    height: 32,
    borderRadius: radius.md,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.xs,
  },
  statAccentDot: { fontSize: 14 },
  statLabel: { fontSize: 11, color: colors.textSecondary, fontWeight: '500' },
  statValue: { fontSize: 20, fontWeight: '700', color: colors.textPrimary },
  statSub: { fontSize: 11, color: colors.textSecondary },
});
