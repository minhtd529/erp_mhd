import React from 'react';
import { View, Text, ScrollView, StyleSheet } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import { SafeAreaView } from 'react-native-safe-area-context';
import { TouchableOpacity } from 'react-native';
import { Card } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { PageSpinner } from '@/components/ui/Spinner';
import { engagementService, STATUS_LABELS, type EngagementStatus } from '@/services/engagements';
import { formatCurrency, formatDate } from '@/lib/format';
import { colors, spacing, typography } from '@/lib/theme';

const STATUS_VARIANTS: Record<EngagementStatus, 'ghost' | 'warning' | 'default' | 'success' | 'secondary'> = {
  DRAFT: 'ghost', PROPOSAL: 'warning', CONTRACTED: 'default',
  ACTIVE: 'success', COMPLETED: 'secondary', SETTLED: 'ghost',
};

function Row({ label, value }: { label: string; value?: string | null }) {
  if (!value) return null;
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

export default function EngagementDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();

  const { data, isLoading } = useQuery({
    queryKey: ['engagement', id],
    queryFn: () => engagementService.get(id),
    enabled: !!id,
  });

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backBtn}>
          <Text style={styles.backText}>← Quay lại</Text>
        </TouchableOpacity>
        <Text style={styles.headerTitle}>Chi tiết hợp đồng</Text>
      </View>

      {isLoading ? <PageSpinner /> : data ? (
        <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
          <Card>
            <View style={styles.titleRow}>
              <View style={{ flex: 1 }}>
                <Text style={typography.h3}>{data.title}</Text>
                <Text style={styles.code}>{data.code}</Text>
              </View>
              <Badge label={STATUS_LABELS[data.status]} variant={STATUS_VARIANTS[data.status]} />
            </View>
            {data.description && (
              <Text style={[typography.small, { marginTop: spacing.sm }]}>{data.description}</Text>
            )}
          </Card>

          <Card>
            <Text style={[typography.label, { marginBottom: spacing.md }]}>Thông tin chung</Text>
            <Row label="Khách hàng" value={data.client_name} />
            <Row label="Loại dịch vụ" value={data.service_type} />
            <Row label="Ngân sách" value={data.budget ? formatCurrency(data.budget) : null} />
            <Row label="Ngày bắt đầu" value={data.start_date ? formatDate(data.start_date) : null} />
            <Row label="Ngày kết thúc" value={data.end_date ? formatDate(data.end_date) : null} />
            <Row label="Cập nhật" value={formatDate(data.updated_at)} />
          </Card>

          <View style={styles.timeline}>
            <Text style={[typography.label, { marginBottom: spacing.md }]}>Trạng thái hợp đồng</Text>
            {(['DRAFT', 'PROPOSAL', 'CONTRACTED', 'ACTIVE', 'COMPLETED', 'SETTLED'] as EngagementStatus[]).map((status, i) => {
              const statusOrder = ['DRAFT', 'PROPOSAL', 'CONTRACTED', 'ACTIVE', 'COMPLETED', 'SETTLED'];
              const currentIdx = statusOrder.indexOf(data.status);
              const thisIdx = statusOrder.indexOf(status);
              const isActive = data.status === status;
              const isDone = thisIdx < currentIdx;
              return (
                <View key={status} style={styles.timelineItem}>
                  <View style={[
                    styles.timelineDot,
                    isActive && styles.timelineDotActive,
                    isDone && styles.timelineDotDone,
                  ]} />
                  {i < 5 && <View style={[styles.timelineLine, (isDone || isActive) && styles.timelineLineDone]} />}
                  <Text style={[
                    styles.timelineLabel,
                    isActive && { color: colors.primary, fontWeight: '600' },
                    isDone && { color: colors.success },
                  ]}>
                    {STATUS_LABELS[status]}
                  </Text>
                </View>
              );
            })}
          </View>
        </ScrollView>
      ) : (
        <View style={styles.errorContainer}>
          <Text style={typography.small}>Không tìm thấy hợp đồng</Text>
        </View>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    gap: spacing.md,
  },
  backBtn: { paddingVertical: spacing.xs },
  backText: { fontSize: 14, color: colors.primaryLight, fontWeight: '500' },
  headerTitle: { fontSize: 16, fontWeight: '600', color: colors.textPrimary },
  content: { padding: spacing.lg, gap: spacing.md },
  titleRow: { flexDirection: 'row', gap: spacing.md, alignItems: 'flex-start' },
  code: { fontSize: 11, color: colors.textSecondary, fontFamily: 'monospace', marginTop: 2 },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: spacing.sm, borderBottomWidth: 1, borderBottomColor: colors.border },
  rowLabel: { fontSize: 13, color: colors.textSecondary },
  rowValue: { fontSize: 13, color: colors.textPrimary, fontWeight: '500', maxWidth: '55%', textAlign: 'right' },
  timeline: { backgroundColor: colors.surface, borderRadius: 8, padding: spacing.lg, borderWidth: 1, borderColor: colors.border },
  timelineItem: { flexDirection: 'row', alignItems: 'flex-start', marginBottom: 4 },
  timelineDot: { width: 10, height: 10, borderRadius: 5, backgroundColor: colors.border, marginTop: 3, marginRight: spacing.md },
  timelineDotActive: { backgroundColor: colors.primary, width: 12, height: 12, borderRadius: 6 },
  timelineDotDone: { backgroundColor: colors.success },
  timelineLine: { position: 'absolute', left: 4.5, top: 13, width: 1, height: 20, backgroundColor: colors.border },
  timelineLineDone: { backgroundColor: colors.success },
  timelineLabel: { fontSize: 13, color: colors.textSecondary, flex: 1 },
  errorContainer: { flex: 1, alignItems: 'center', justifyContent: 'center' },
});
