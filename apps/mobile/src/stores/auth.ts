import { create } from 'zustand';
import * as SecureStore from 'expo-secure-store';

export interface AuthUser {
  id: string;
  email: string;
  full_name: string;
  roles: string[];
  two_factor_enabled: boolean;
}

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  pendingChallengeId: string | null;
  hydrated: boolean;
  hydrate: () => Promise<void>;
  setTokens: (access: string, refresh: string) => Promise<void>;
  setUser: (user: AuthUser) => void;
  setPendingChallenge: (id: string) => void;
  logout: () => Promise<void>;
  isAuthenticated: () => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  refreshToken: null,
  pendingChallengeId: null,
  hydrated: false,

  hydrate: async () => {
    const access = await SecureStore.getItemAsync('access_token');
    const refresh = await SecureStore.getItemAsync('refresh_token');
    const userJson = await SecureStore.getItemAsync('user');
    const user = userJson ? JSON.parse(userJson) as AuthUser : null;
    set({ accessToken: access, refreshToken: refresh, user, hydrated: true });
  },

  setTokens: async (access, refresh) => {
    await SecureStore.setItemAsync('access_token', access);
    await SecureStore.setItemAsync('refresh_token', refresh);
    set({ accessToken: access, refreshToken: refresh });
  },

  setUser: (user) => {
    SecureStore.setItemAsync('user', JSON.stringify(user));
    set({ user });
  },

  setPendingChallenge: (id) => set({ pendingChallengeId: id }),

  logout: async () => {
    await SecureStore.deleteItemAsync('access_token');
    await SecureStore.deleteItemAsync('refresh_token');
    await SecureStore.deleteItemAsync('user');
    set({ user: null, accessToken: null, refreshToken: null, pendingChallengeId: null });
  },

  isAuthenticated: () => !!get().accessToken && !!get().user,

  hasRole: (role) => get().user?.roles?.includes(role) ?? false,
}));
