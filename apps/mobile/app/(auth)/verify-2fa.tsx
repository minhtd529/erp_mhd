import React from 'react';
import {
  View, Text, StyleSheet, TouchableOpacity, Alert, KeyboardAvoidingView, Platform,
} from 'react-native';
import { useRouter } from 'expo-router';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { authService } from '@/services/auth';
import { useAuthStore } from '@/stores/auth';
import { getErrorMessage } from '@/lib/api';
import { colors, spacing, radius, typography } from '@/lib/theme';

export default function Verify2FAScreen() {
  const router = useRouter();
  const { pendingChallengeId, setTokens, setUser } = useAuthStore();
  const [code, setCode] = React.useState('');
  const [useBackup, setUseBackup] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (!pendingChallengeId) router.replace('/(auth)/login');
  }, [pendingChallengeId]);

  const handleVerify = async () => {
    if (!pendingChallengeId || !code.trim()) return;
    setLoading(true);
    try {
      const res = useBackup
        ? await authService.verifyBackupCode(pendingChallengeId, code.trim())
        : await authService.verify2FA(pendingChallengeId, code.trim());
      await setTokens(res.access_token, res.refresh_token);
      const me = await authService.me();
      setUser(me);
      router.replace('/(app)/dashboard');
    } catch (err) {
      Alert.alert('Xác thực thất bại', getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
        <View style={styles.container}>
          <View style={styles.icon}>
            <Text style={styles.iconText}>🔐</Text>
          </View>

          <Text style={typography.h2}>Xác thực 2 lớp</Text>
          <Text style={[typography.small, { textAlign: 'center' }]}>
            {useBackup
              ? 'Nhập backup code (8 ký tự)'
              : 'Nhập mã 6 chữ số từ ứng dụng authenticator'}
          </Text>

          <Input
            value={code}
            onChangeText={setCode}
            placeholder={useBackup ? 'XXXXXXXX' : '000000'}
            keyboardType="number-pad"
            maxLength={useBackup ? 8 : 6}
            style={styles.codeInput}
            autoFocus
          />

          <Button
            title="Xác thực"
            onPress={handleVerify}
            loading={loading}
            disabled={!code.trim()}
            style={styles.btn}
          />

          <TouchableOpacity onPress={() => { setUseBackup(v => !v); setCode(''); }}>
            <Text style={styles.linkText}>
              {useBackup ? '← Dùng mã TOTP' : 'Dùng backup code →'}
            </Text>
          </TouchableOpacity>

          <TouchableOpacity onPress={() => router.replace('/(auth)/login')}>
            <Text style={styles.backText}>← Quay lại đăng nhập</Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  flex: { flex: 1 },
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.xxl,
    gap: spacing.lg,
  },
  icon: {
    width: 72, height: 72, borderRadius: radius.xl,
    backgroundColor: colors.primary,
    alignItems: 'center', justifyContent: 'center',
    marginBottom: spacing.sm,
  },
  iconText: { fontSize: 32 },
  codeInput: {
    textAlign: 'center',
    fontSize: 24,
    letterSpacing: 8,
    fontWeight: '700',
    width: '100%',
    height: 60,
  },
  btn: { width: '100%' },
  linkText: { fontSize: 13, color: colors.primaryLight, fontWeight: '500' },
  backText: { fontSize: 12, color: colors.textSecondary },
});
