'use client';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { authService } from '@/services/auth';
import { useAuthStore } from '@/stores/auth';
import { getErrorMessage } from '@/lib/utils';
import { ShieldCheck } from 'lucide-react';
import { getRoleLandingPage } from '@/lib/roles';

export default function Verify2FAPage() {
  const router = useRouter();
  const { pendingChallengeId, setTokens, setUser } = useAuthStore();
  const [code, setCode] = React.useState('');
  const [useBackup, setUseBackup] = React.useState(false);
  const [error, setError] = React.useState('');
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (!pendingChallengeId) router.replace('/login');
  }, [pendingChallengeId, router]);

  const handleVerify = async () => {
    if (!pendingChallengeId || !code) return;
    setError('');
    setLoading(true);
    try {
      const res = useBackup
        ? await authService.verifyBackupCode(pendingChallengeId, code)
        : await authService.verify2FA({ challenge_id: pendingChallengeId, code });
      if (res.access_token && res.refresh_token) {
        setTokens(res.access_token, res.refresh_token);
        const me = await authService.me();
        setUser(me);
        router.push(getRoleLandingPage(me.roles ?? []));
      }
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="w-full max-w-sm">
        <div className="flex justify-center mb-8">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-card bg-primary flex items-center justify-center">
              <ShieldCheck className="w-6 h-6 text-white" />
            </div>
            <p className="text-sm font-semibold text-primary">ERP Audit</p>
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Xác thực 2 lớp</CardTitle>
            <CardDescription>
              {useBackup ? 'Nhập backup code' : 'Nhập mã TOTP từ ứng dụng authenticator'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label>{useBackup ? 'Backup Code' : 'Mã xác thực (6 chữ số)'}</Label>
                <Input
                  value={code}
                  onChange={e => setCode(e.target.value)}
                  placeholder={useBackup ? 'XXXXXXXX' : '000000'}
                  maxLength={useBackup ? 8 : 6}
                  className="text-center text-lg tracking-widest font-mono"
                  onKeyDown={e => e.key === 'Enter' && handleVerify()}
                  autoFocus
                />
              </div>

              {error && (
                <div className="rounded bg-red-50 border border-danger/20 px-3 py-2 text-sm text-danger">
                  {error}
                </div>
              )}

              <Button onClick={handleVerify} loading={loading} className="w-full">
                Xác thực
              </Button>

              <button
                type="button"
                className="text-xs text-action hover:underline text-center"
                onClick={() => { setUseBackup(v => !v); setCode(''); setError(''); }}
              >
                {useBackup ? 'Dùng mã TOTP' : 'Dùng backup code'}
              </button>

              <button
                type="button"
                className="text-xs text-text-secondary hover:underline text-center"
                onClick={() => router.push('/login')}
              >
                ← Quay lại đăng nhập
              </button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
