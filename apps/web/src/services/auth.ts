import { api } from '@/lib/api';

export interface LoginRequest { email: string; password: string; }
export interface LoginResponse {
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
  challenge_id?: string;
  challenge_type?: 'totp' | 'push';
}
export interface Verify2FARequest { challenge_id: string; code: string; }
export interface MeResponse {
  id: string; email: string; full_name: string;
  roles: string[]; two_factor_enabled: boolean;
  branch_id?: string; department_id?: string;
  permissions?: string[];
}

export const authService = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data).then(r => r.data),
  verify2FA: (data: Verify2FARequest) => api.post<LoginResponse>('/auth/2fa/verify', data).then(r => r.data),
  verifyBackupCode: (challenge_id: string, code: string) =>
    api.post<LoginResponse>('/auth/2fa/backup', { challenge_id, code }).then(r => r.data),
  me: () => api.get<MeResponse>('/me').then(r => r.data),
  logout: () => api.post('/auth/logout').catch(() => {}),
  refreshToken: (token: string) =>
    api.post<{ access_token: string; refresh_token: string }>('/auth/refresh', { refresh_token: token }).then(r => r.data),
};
