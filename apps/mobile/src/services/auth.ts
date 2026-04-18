import { api } from '@/lib/api';

export const authService = {
  login: (email: string, password: string) =>
    api.post<{ access_token?: string; refresh_token?: string; challenge_id?: string; challenge_type?: string }>(
      '/auth/login', { email, password }
    ).then(r => r.data),

  verify2FA: (challenge_id: string, code: string) =>
    api.post<{ access_token: string; refresh_token: string }>(
      '/auth/2fa/verify', { challenge_id, code }
    ).then(r => r.data),

  verifyBackupCode: (challenge_id: string, code: string) =>
    api.post<{ access_token: string; refresh_token: string }>(
      '/auth/2fa/backup', { challenge_id, code }
    ).then(r => r.data),

  me: () =>
    api.get<{ id: string; email: string; full_name: string; roles: string[]; two_factor_enabled: boolean }>(
      '/auth/me'
    ).then(r => r.data),

  logout: () => api.post('/auth/logout').catch(() => {}),
};
