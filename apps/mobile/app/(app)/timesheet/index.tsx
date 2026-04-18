import React from 'react';
import { View, Text, FlatList, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Badge } from '@/components/ui/Badge';
import { ListItem } from '@/components/ui/ListItem';
import { PageSpinner } from '@/components/ui/Spinner';
import { EmptyState } from '@/components/ui/EmptyState';
import { timesheetService, STATUS_LABELS, type TimesheetStatus } from '@/services/timesheets';
import { getErrorMessage } from '@/lib/api';
import { formatDate } from '@/lib/format';
import { colors, spacing, typography } from '@/lib/theme';

const STATUS_VARIANTS: Record<TimesheetStatus, 'ghost' | 'warning' | 'success' | 'danger' | 'secondary'> = {
  OPEN: 'ghost', SUBMITTED: 'warning', APPROVED: 'success', REJECTED: 'danger', LOCKED: 'secondary',
};

export default function TimesheetScreen() {
  const router = useRouter();
  const qc = useQueryClient();

  const { data, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['timesheets-mobile'],
    queryFn: () => timesheetService.list({ size: 30 }),
  });

  const submitMut = useMutation({
    mutationFn: (id: string) => timesheetService.submit(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['timesheets-mobile'] }),
    onError: (err) => Alert.alert('Lỗi', getErrorMessage(err)),
  });

  const handleSubmit = (id: string) => {
    Alert.alert('Gửi duyệt', 'Bạn có chắc muốn gửi chấm công này để duyệt?', [
      { text: 'Hủy', style: 'cancel' },
      { text: 'Gửi', onPress: () => submitMut.mutate(id) },
    ]);
  };

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.title}>Chấm công</Text>
      </View>

      {isLoading ? <PageSpinner /> : (
        <FlatList
          data={data?.data ?? []}
          keyExtractor={item => item.id}
          onRefresh={refetch}
          refreshing={isRefetching}
          ListEmptyComponent={<EmptyState message="Không có dữ liệu chấm công" icon="⏱" />}
          contentContainerStyle={data?.data.length === 0 ? { flex: 1 } : undefined}
          renderItem={({ item }) => (
            <View>
              <ListItem
                title={`${formatDate(item.period_start)} — ${formatDate(item.period_end)}`}
                subtitle={item.total_hours ? `${item.total_hours}h ghi nhận` : 'Chưa có giờ'}
                onPress={() => router.push(`/(app)/timesheet/entries?id=${item.id}`)}
                right={
                  <View style={{ alignItems: 'flex-end', gap: spacing.xs }}>
                    <Badge label={STATUS_LABELS[item.status]} variant={STATUS_VARIANTS[item.status]} />
                    {item.status === 'OPEN' && (
                      <TouchableOpacity
                        onPress={() => handleSubmit(item.id)}
                        style={styles.submitBtn}
                      >
                        <Text style={styles.submitText}>Gửi duyệt</Text>
                      </TouchableOpacity>
                    )}
                  </View>
                }
              />
            </View>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  header: {
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.md,
    paddingBottom: spacing.md,
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  title: { fontSize: 18, fontWeight: '700', color: colors.textPrimary },
  submitBtn: {
    backgroundColor: colors.primary,
    borderRadius: 4,
    paddingHorizontal: spacing.sm,
    paddingVertical: 3,
  },
  submitText: { fontSize: 11, color: '#fff', fontWeight: '600' },
});
