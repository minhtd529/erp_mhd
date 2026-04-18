import React from 'react';
import { View, Text, ScrollView, StyleSheet, RefreshControl } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Card, StatCard } from '@/components/ui/Card';
import { PageSpinner } from '@/components/ui/Spinner';
import { Badge } from '@/components/ui/Badge';
import { reportService } from '@/services/reports';
import { useAuthStore } from '@/stores/auth';
import { formatCurrency } from '@/lib/format';
import { colors, spacing, typography } from '@/lib/theme';

function CommissionWidget({ data }: { data: NonNullable<ReturnType<typeof reportService.personal> extends Promise<infer T> ? T : never> }) {
  if (!data.is_salesperson) return null;
  return (
    <Card style={styles.commissionCard}>
      <View style={styles.commissionHeader}>
        <Text style={typography.h3}>Hoa hồng của tôi</Text>
        <Badge label="Salesperson" variant="success" />
      </View>
      <View style={styles.commissionGrid}>
        {[
          { label: 'Cả năm (YTD)', value: data.commission_ytd ?? 0, color: colors.primary },
          { label: 'Tháng này', value: data.commission_month ?? 0, color: colors.success },
          { label: 'Chờ duyệt', value: data.commission_pending ?? 0, color: colors.warning },
          { label: 'Đang giữ', value: data.commission_on_hold ?? 0, color: colors.textSecondary },
        ].map(item => (
          <View key={item.label} style={styles.commissionItem}>
            <Text style={styles.commissionLabel}>{item.label}</Text>
            <Text style={[styles.commissionValue, { color: item.color }]}>
              {formatCurrency(item.value)}
            </Text>
          </View>
        ))}
      </View>
    </Card>
  );
}

export default function DashboardScreen() {
  const { user } = useAuthStore();
  const { data, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['personal-dashboard'],
    queryFn: reportService.personal,
  });

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <View>
          <Text style={typography.h2}>Xin chào, {user?.full_name?.split(' ').pop() ?? 'bạn'} 👋</Text>
          <Text style={typography.small}>Tổng quan hoạt động hôm nay</Text>
        </View>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>{user?.full_name?.[0]?.toUpperCase() ?? 'U'}</Text>
        </View>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={<RefreshControl refreshing={isRefetching} onRefresh={refetch} tintColor={colors.primary} />}
      >
        {isLoading ? (
          <PageSpinner />
        ) : data ? (
          <>
            <View style={styles.statsRow}>
              <StatCard label="Hợp đồng đang tham gia" value={String(data.active_engagements)} />
              <StatCard label="Chấm công chờ duyệt" value={String(data.pending_timesheets)} accent="#FEF3C7" />
            </View>
            <View style={styles.statsRow}>
              <StatCard label="Giờ làm tháng này" value={`${data.total_hours_this_month}h`} accent={colors.successLight} />
              <Card style={{ flex: 1 }}>
                <Text style={typography.small}>Vai trò</Text>
                <View style={{ marginTop: spacing.sm, gap: spacing.xs, flexDirection: 'row', flexWrap: 'wrap' }}>
                  {user?.roles?.map(r => (
                    <Badge key={r} label={r.replace('_', ' ')} variant="default" />
                  ))}
                </View>
              </Card>
            </View>
            <CommissionWidget data={data} />
          </>
        ) : (
          <Card>
            <Text style={[typography.small, { textAlign: 'center' }]}>Không thể tải dữ liệu</Text>
          </Card>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.md,
    paddingBottom: spacing.lg,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  avatar: {
    width: 40, height: 40, borderRadius: 20,
    backgroundColor: colors.primary,
    alignItems: 'center', justifyContent: 'center',
  },
  avatarText: { fontSize: 16, fontWeight: '700', color: '#fff' },
  scroll: { flex: 1 },
  content: { padding: spacing.lg, gap: spacing.md },
  statsRow: { flexDirection: 'row', gap: spacing.md },
  commissionCard: { gap: spacing.md },
  commissionHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  commissionGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  commissionItem: { width: '45%', gap: 4 },
  commissionLabel: { fontSize: 11, color: colors.textSecondary, fontWeight: '500' },
  commissionValue: { fontSize: 15, fontWeight: '700' },
});
