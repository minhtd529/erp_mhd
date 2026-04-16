// ─── Common ─────────────────────────────────────────────────────────────────

export interface PaginatedResult<T> {
  data: T[];
  total: number;
  page: number;
  size: number;
  total_pages: number;
}

export interface ApiError {
  error: string;   // UPPER_SNAKE_CASE error code
  message: string; // Human-readable message
}

// ─── Auth ────────────────────────────────────────────────────────────────────

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token?: string;
  refresh_token?: string;
  expires_in?: number;
  challenge_id?: string;
  challenge_type?: 'totp' | 'push';
}

export interface User {
  id: string;
  email: string;
  full_name: string;
  status: 'active' | 'inactive' | 'locked';
  two_factor_enabled: boolean;
  created_at: string;
  updated_at: string;
}
