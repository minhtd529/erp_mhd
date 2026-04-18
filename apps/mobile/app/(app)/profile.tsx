import React from 'react';
import { View, Text, StyleSheet, Alert, ScrollView, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useAuthStore } from '@/stores/auth';
import { authService } from '@/services/auth';
import { useRouter } from 'expo-router';
import { colors, spacing, typography } from '@/lib/theme';

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

export default function ProfileScreen() {
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = () => {
    Alert.alert('Đăng xuất', 'Bạn có chắc muốn đăng xuất?', [
      { text: 'Hủy', style: 'cancel' },
      {
        text: 'Đăng xuất',
        style: 'destructive',
        onPress: async () => {
          await authService.logout();
          await logout();
          router.replace('/(auth)/login');
        },
      },
    ]);
  };

  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.title}>Tài khoản</Text>
      </View>

      <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}>
        <View style={styles.avatarSection}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{user?.full_name?.[0]?.toUpperCase() ?? 'U'}</Text>
          </View>
          <Text style={typography.h3}>{user?.full_name}</Text>
          <Text style={typography.small}>{user?.email}</Text>
        </View>

        <Card>
          <Text style={[typography.label, { marginBottom: spacing.md }]}>Thông tin tài khoản</Text>
          <InfoRow label="Email" value={user?.email ?? '-'} />
          <InfoRow label="2FA" value={user?.two_factor_enabled ? 'Đã bật' : 'Chưa bật'} />
        </Card>

        <Card>
          <Text style={[typography.label, { marginBottom: spacing.md }]}>Vai trò</Text>
          <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm }}>
            {user?.roles?.map(role => (
              <Badge key={role} label={role.replace('_', ' ')} variant="default" />
            ))}
          </View>
        </Card>

        <Card>
          <Text style={[typography.label, { marginBottom: spacing.md }]}>Thông tin ứng dụng</Text>
          <InfoRow label="Phiên bản" value="1.0.0" />
          <InfoRow label="Hệ thống" value="ERP Audit System" />
        </Card>

        <Button
          title="Đăng xuất"
          variant="danger"
          onPress={handleLogout}
        />
      </ScrollView>
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
  content: { padding: spacing.lg, gap: spacing.md },
  avatarSection: { alignItems: 'center', gap: spacing.sm, paddingVertical: spacing.lg },
  avatar: {
    width: 72, height: 72, borderRadius: 36,
    backgroundColor: colors.primary,
    alignItems: 'center', justifyContent: 'center',
  },
  avatarText: { fontSize: 28, fontWeight: '700', color: '#fff' },
  row: {
    flexDirection: 'row', justifyContent: 'space-between',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1, borderBottomColor: colors.border,
  },
  rowLabel: { fontSize: 13, color: colors.textSecondary },
  rowValue: { fontSize: 13, color: colors.textPrimary, fontWeight: '500' },
});
