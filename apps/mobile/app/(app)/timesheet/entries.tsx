import React from 'react';
import {
  View, Text, FlatList, StyleSheet, TouchableOpacity, Alert,
  Modal, ScrollView, KeyboardAvoidingView, Platform,
} from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { ListItem } from '@/components/ui/ListItem';
import { PageSpinner } from '@/components/ui/Spinner';
import { EmptyState } from '@/components/ui/EmptyState';
import { timesheetService } from '@/services/timesheets';
import { engagementService } from '@/services/engagements';
import { getErrorMessage } from '@/lib/api';
import { formatDate } from '@/lib/format';
import { colors, spacing, typography } from '@/lib/theme';

const schema = z.object({
  engagement_id: z.string().min(1, 'Chọn hợp đồng'),
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Định dạng YYYY-MM-DD'),
  hours: z.coerce.number().min(0.5, 'Tối thiểu 0.5h').max(24, 'Tối đa 24h'),
  description: z.string().optional(),
});
type FormData = z.infer<typeof schema>;

export default function TimesheetEntriesScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const [showModal, setShowModal] = React.useState(false);

  const { data: entries, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['timesheet-entries', id],
    queryFn: () => timesheetService.listEntries(id),
    enabled: !!id,
  });

  const { data: engagements } = useQuery({
    queryKey: ['engagements-active'],
    queryFn: () => engagementService.list({ status: 'ACTIVE', size: 50 }),
  });

  const { control, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { date: new Date().toISOString().split('T')[0], hours: 8 },
  });

  const createMut = useMutation({
    mutationFn: (data: FormData) => timesheetService.createEntry(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['timesheet-entries', id] });
      setShowModal(false);
      reset();
    },
    onError: (err) => Alert.alert('Lỗi', getErrorMessage(err)),
  });

  const deleteMut = useMutation({
    mutationFn: (entryId: string) => timesheetService.deleteEntry(id, entryId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['timesheet-entries', id] }),
    onError: (err) => Alert.alert('Lỗi', getErrorMessage(err)),
  });

  const totalHours = entries?.data.reduce((sum, e) => sum + e.hours, 0) ?? 0;

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()}>
          <Text style={styles.back}>← Quay lại</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Ghi giờ</Text>
        <TouchableOpacity onPress={() => setShowModal(true)} style={styles.addBtn}>
          <Text style={styles.addText}>+ Thêm</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.summary}>
        <Text style={typography.small}>Tổng giờ</Text>
        <Text style={styles.totalHours}>{totalHours}h</Text>
      </View>

      {isLoading ? <PageSpinner /> : (
        <FlatList
          data={entries?.data ?? []}
          keyExtractor={item => item.id}
          onRefresh={refetch}
          refreshing={isRefetching}
          ListEmptyComponent={<EmptyState message="Chưa có giờ nào được ghi" icon="⏰" />}
          contentContainerStyle={entries?.data.length === 0 ? { flex: 1 } : undefined}
          renderItem={({ item }) => (
            <ListItem
              title={`${item.hours}h — ${item.engagement_title ?? item.engagement_id.slice(0, 8)}`}
              subtitle={`${formatDate(item.date)}${item.description ? ` · ${item.description}` : ''}`}
              right={
                <TouchableOpacity
                  onPress={() => Alert.alert('Xóa?', '', [
                    { text: 'Hủy', style: 'cancel' },
                    { text: 'Xóa', style: 'destructive', onPress: () => deleteMut.mutate(item.id) },
                  ])}
                >
                  <Text style={{ fontSize: 18, color: colors.danger }}>×</Text>
                </TouchableOpacity>
              }
            />
          )}
        />
      )}

      <Modal visible={showModal} animationType="slide" presentationStyle="pageSheet">
        <SafeAreaView style={styles.modalSafe}>
          <KeyboardAvoidingView
            style={{ flex: 1 }}
            behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          >
            <View style={styles.modalHeader}>
              <Text style={typography.h3}>Thêm giờ làm</Text>
              <TouchableOpacity onPress={() => { setShowModal(false); reset(); }}>
                <Text style={{ fontSize: 22, color: colors.textSecondary }}>×</Text>
              </TouchableOpacity>
            </View>
            <ScrollView contentContainerStyle={styles.modalContent} keyboardShouldPersistTaps="handled">
              <Controller
                control={control}
                name="engagement_id"
                render={({ field: { onChange, value } }) => (
                  <View>
                    <Text style={typography.label}>Hợp đồng *</Text>
                    <ScrollView
                      horizontal
                      style={styles.engagementScroll}
                      showsHorizontalScrollIndicator={false}
                    >
                      {engagements?.data.map(eng => (
                        <TouchableOpacity
                          key={eng.id}
                          style={[styles.engChip, value === eng.id && styles.engChipActive]}
                          onPress={() => onChange(eng.id)}
                        >
                          <Text style={[styles.engChipText, value === eng.id && { color: '#fff' }]}>
                            {eng.title}
                          </Text>
                        </TouchableOpacity>
                      ))}
                    </ScrollView>
                    {errors.engagement_id && <Text style={styles.errorText}>{errors.engagement_id.message}</Text>}
                  </View>
                )}
              />

              <Controller
                control={control}
                name="date"
                render={({ field: { onChange, value } }) => (
                  <Input
                    label="Ngày (YYYY-MM-DD) *"
                    value={value}
                    onChangeText={onChange}
                    placeholder="2026-04-18"
                    error={errors.date?.message}
                  />
                )}
              />

              <Controller
                control={control}
                name="hours"
                render={({ field: { onChange, value } }) => (
                  <Input
                    label="Số giờ *"
                    value={String(value)}
                    onChangeText={onChange}
                    keyboardType="decimal-pad"
                    placeholder="8"
                    error={errors.hours?.message}
                  />
                )}
              />

              <Controller
                control={control}
                name="description"
                render={({ field: { onChange, value } }) => (
                  <Input
                    label="Mô tả công việc"
                    value={value}
                    onChangeText={onChange}
                    placeholder="Soát xét BCTC, kiểm tra hồ sơ..."
                    error={errors.description?.message}
                  />
                )}
              />

              <Button
                title="Lưu"
                onPress={handleSubmit(d => createMut.mutate(d))}
                loading={createMut.isPending || isSubmitting}
                style={{ marginTop: spacing.md }}
              />
            </ScrollView>
          </KeyboardAvoidingView>
        </SafeAreaView>
      </Modal>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  header: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between',
    paddingHorizontal: spacing.lg, paddingVertical: spacing.md,
    backgroundColor: colors.surface, borderBottomWidth: 1, borderBottomColor: colors.border,
  },
  back: { fontSize: 14, color: colors.primaryLight, fontWeight: '500' },
  title: { fontSize: 16, fontWeight: '700', color: colors.textPrimary },
  addBtn: { backgroundColor: colors.primary, borderRadius: 6, paddingHorizontal: spacing.md, paddingVertical: 6 },
  addText: { fontSize: 13, color: '#fff', fontWeight: '600' },
  summary: {
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: spacing.lg, paddingVertical: spacing.md,
    backgroundColor: colors.surface, borderBottomWidth: 1, borderBottomColor: colors.border,
  },
  totalHours: { fontSize: 18, fontWeight: '700', color: colors.primary },
  modalSafe: { flex: 1, backgroundColor: colors.surface },
  modalHeader: {
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: spacing.lg, paddingVertical: spacing.md,
    borderBottomWidth: 1, borderBottomColor: colors.border,
  },
  modalContent: { padding: spacing.lg, gap: spacing.md },
  engagementScroll: { marginTop: spacing.sm, marginBottom: spacing.xs },
  engChip: {
    paddingHorizontal: spacing.md, paddingVertical: 8,
    borderRadius: 20, borderWidth: 1, borderColor: colors.border,
    marginRight: spacing.sm, backgroundColor: colors.surface,
  },
  engChipActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  engChipText: { fontSize: 13, color: colors.textPrimary },
  errorText: { fontSize: 12, color: colors.danger, marginTop: 2 },
});
