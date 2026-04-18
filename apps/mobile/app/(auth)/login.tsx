import React from 'react';
import {
  View, Text, StyleSheet, ScrollView, KeyboardAvoidingView, Platform,
  TouchableOpacity, Alert,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { authService } from '@/services/auth';
import { useAuthStore } from '@/stores/auth';
import { getErrorMessage } from '@/lib/api';
import { colors, spacing, radius, typography } from '@/lib/theme';

const schema = z.object({
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(1, 'Nhập mật khẩu'),
});
type FormData = z.infer<typeof schema>;

export default function LoginScreen() {
  const router = useRouter();
  const { setTokens, setUser, setPendingChallenge } = useAuthStore();
  const [showPw, setShowPw] = React.useState(false);

  const { control, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: FormData) => {
    try {
      const res = await authService.login(data.email, data.password);
      if (res.challenge_id) {
        setPendingChallenge(res.challenge_id);
        router.push('/(auth)/verify-2fa');
        return;
      }
      if (res.access_token && res.refresh_token) {
        await setTokens(res.access_token, res.refresh_token);
        const me = await authService.me();
        setUser(me);
        router.replace('/(app)/dashboard');
      }
    } catch (err) {
      Alert.alert('Đăng nhập thất bại', getErrorMessage(err));
    }
  };

  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <ScrollView
          contentContainerStyle={styles.scroll}
          keyboardShouldPersistTaps="handled"
        >
          <View style={styles.brand}>
            <View style={styles.logo}>
              <Text style={styles.logoText}>⚖</Text>
            </View>
            <Text style={styles.brandName}>ERP Audit</Text>
            <Text style={styles.brandSub}>Hệ thống quản lý kiểm toán</Text>
          </View>

          <View style={styles.form}>
            <Text style={typography.h2}>Đăng nhập</Text>
            <Text style={[typography.small, { marginBottom: spacing.lg }]}>
              Nhập thông tin tài khoản của bạn
            </Text>

            <Controller
              control={control}
              name="email"
              render={({ field: { onChange, value } }) => (
                <Input
                  label="Email"
                  value={value}
                  onChangeText={onChange}
                  placeholder="name@firm.com"
                  keyboardType="email-address"
                  autoCapitalize="none"
                  autoCorrect={false}
                  error={errors.email?.message}
                />
              )}
            />

            <View style={{ gap: 4, marginTop: spacing.md }}>
              <Controller
                control={control}
                name="password"
                render={({ field: { onChange, value } }) => (
                  <Input
                    label="Mật khẩu"
                    value={value}
                    onChangeText={onChange}
                    placeholder="••••••••"
                    secureTextEntry={!showPw}
                    autoCapitalize="none"
                    error={errors.password?.message}
                  />
                )}
              />
              <TouchableOpacity onPress={() => setShowPw(v => !v)} style={styles.showPw}>
                <Text style={styles.showPwText}>{showPw ? 'Ẩn' : 'Hiện'} mật khẩu</Text>
              </TouchableOpacity>
            </View>

            <Button
              title="Đăng nhập"
              onPress={handleSubmit(onSubmit)}
              loading={isSubmitting}
              style={{ marginTop: spacing.xl }}
            />
          </View>

          <Text style={styles.footer}>ERP Audit System — v1.0</Text>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  flex: { flex: 1 },
  scroll: { flexGrow: 1, justifyContent: 'center', padding: spacing.xl, gap: spacing.xxxl },
  brand: { alignItems: 'center', gap: spacing.sm },
  logo: {
    width: 64, height: 64, borderRadius: radius.lg,
    backgroundColor: colors.primary,
    alignItems: 'center', justifyContent: 'center',
  },
  logoText: { fontSize: 28, color: '#fff' },
  brandName: { fontSize: 20, fontWeight: '700', color: colors.primary },
  brandSub: { fontSize: 13, color: colors.textSecondary },
  form: {
    backgroundColor: colors.surface,
    borderRadius: radius.xl,
    padding: spacing.xl,
    borderWidth: 1,
    borderColor: colors.border,
    gap: spacing.xs,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 8,
    elevation: 3,
  },
  showPw: { alignSelf: 'flex-end' },
  showPwText: { fontSize: 12, color: colors.primaryLight },
  footer: { textAlign: 'center', fontSize: 11, color: colors.textSecondary },
});
