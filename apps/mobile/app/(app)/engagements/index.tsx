import React from 'react';
import { View, Text, FlatList, StyleSheet, TextInput, TouchableOpacity } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Badge } from '@/components/ui/Badge';
import { ListItem } from '@/components/ui/ListItem';
import { PageSpinner } from '@/components/ui/Spinner';
import { EmptyState } from '@/components/ui/EmptyState';
import { engagementService, STATUS_LABELS, type EngagementStatus } from '@/services/engagements';
import { formatCurrency, formatDate } from '@/lib/format';
import { colors, spacing } from '@/lib/theme';

const STATUS_VARIANTS: Record<EngagementStatus, 'ghost' | 'warning' | 'default' | 'success' | 'secondary'> = {
  DRAFT: 'ghost', PROPOSAL: 'warning', CONTRACTED: 'default',
  ACTIVE: 'success', COMPLETED: 'secondary', SETTLED: 'ghost',
};

export default function EngagementsScreen() {
  const router = useRouter();
  const [q, setQ] = React.useState('');
  const [debouncedQ, setDebouncedQ] = React.useState('');
  const [page] = React.useState(1);

  React.useEffect(() => {
    const t = setTimeout(() => setDebouncedQ(q), 300);
    return () => clearTimeout(t);
  }, [q]);

  const { data, isLoading, refetch, isRefetching } = useQuery({
    queryKey: ['engagements-mobile', page, debouncedQ],
    queryFn: () => engagementService.list({ page, size: 30, q: debouncedQ || undefined }),
  });

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.title}>Hợp đồng</Text>
      </View>

      <View style={styles.searchBar}>
        <TextInput
          style={styles.searchInput}
          placeholder="🔍  Tìm theo tiêu đề..."
          placeholderTextColor={colors.textSecondary}
          value={q}
          onChangeText={setQ}
          autoCapitalize="none"
        />
      </View>

      {isLoading ? (
        <PageSpinner />
      ) : (
        <FlatList
          data={data?.data ?? []}
          keyExtractor={item => item.id}
          onRefresh={refetch}
          refreshing={isRefetching}
          ListEmptyComponent={<EmptyState message="Không có hợp đồng" icon="📋" />}
          renderItem={({ item }) => (
            <ListItem
              title={item.title}
              subtitle={`${item.client_name ?? ''} · ${item.code}`}
              onPress={() => router.push(`/(app)/engagements/${item.id}`)}
              right={
                <View style={{ alignItems: 'flex-end', gap: 4 }}>
                  <Badge label={STATUS_LABELS[item.status]} variant={STATUS_VARIANTS[item.status]} />
                  {item.budget && (
                    <Text style={styles.budget}>{formatCurrency(item.budget)}</Text>
                  )}
                </View>
              }
            />
          )}
          contentContainerStyle={data?.data.length === 0 ? { flex: 1 } : undefined}
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
  searchBar: {
    backgroundColor: colors.surface,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  searchInput: {
    height: 40,
    backgroundColor: colors.background,
    borderRadius: 20,
    paddingHorizontal: spacing.lg,
    fontSize: 14,
    color: colors.textPrimary,
  },
  budget: { fontSize: 11, color: colors.textSecondary },
});
