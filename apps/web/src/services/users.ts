import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export interface User {
  id: string;
  email: string;
  full_name: string;
  status: 'active' | 'inactive' | 'locked';
  two_factor_enabled: boolean;
  branch_id?: string;
  department_id?: string;
}

export interface UserCreateRequest {
  email: string;
  password: string;
  full_name: string;
  role_code: string;
}

export interface UserUpdateRequest {
  full_name: string;
  status: string;
}

export const userService = {
  list: (params: { page?: number; size?: number; q?: string; status?: string }) =>
    api.get<PaginatedResult<User>>('/users', { params }).then(r => r.data),
  get: (id: string) => api.get<User>(`/users/${id}`).then(r => r.data),
  create: (data: UserCreateRequest) => api.post<User>('/users', data).then(r => r.data),
  update: (id: string, data: UserUpdateRequest) => api.put<User>(`/users/${id}`, data).then(r => r.data),
  delete: (id: string) => api.delete(`/users/${id}`),
  assignRole: (id: string, roleCode: string) =>
    api.post(`/users/${id}/roles`, { role_code: roleCode }).then(r => r.data),
};
