import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface AuthUser {
  id: string;
  email: string;
  full_name: string;
  roles: string[];
  two_factor_enabled: boolean;
  branch_id?: string;
  department_id?: string;
  permissions?: string[];
}

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingChallengeId: string | null;
  setTokens: (access: string, refresh: string) => void;
  setUser: (user: AuthUser) => void;
  setPendingChallenge: (id: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      pendingChallengeId: null,

      setTokens: (access, refresh) => {
        set({ accessToken: access, refreshToken: refresh });
        if (typeof window !== 'undefined') {
          localStorage.setItem('access_token', access);
        }
      },

      setUser: (user) => set({ user }),

      setPendingChallenge: (id) => set({ pendingChallengeId: id }),

      logout: () => {
        set({ user: null, accessToken: null, refreshToken: null, pendingChallengeId: null });
        if (typeof window !== 'undefined') {
          localStorage.removeItem('access_token');
        }
      },

      isAuthenticated: () => !!get().accessToken && !!get().user,

      hasRole: (role) => get().user?.roles?.includes(role) ?? false,
    }),
    {
      name: 'erp-auth',
      partialize: (s) => ({ accessToken: s.accessToken, refreshToken: s.refreshToken, user: s.user }),
    }
  )
);
