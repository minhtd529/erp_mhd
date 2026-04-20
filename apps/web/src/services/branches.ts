import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export interface Branch {
  id: string;
  code: string;
  name: string;
  address?: string;
  phone?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Department {
  id: string;
  branch_id?: string;
  code: string;
  name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface BranchCreateRequest {
  code: string;
  name: string;
  address?: string;
  phone?: string;
}

export interface BranchUpdateRequest {
  code: string;
  name: string;
  address?: string;
  phone?: string;
  is_active: boolean;
}

export interface DepartmentCreateRequest {
  code: string;
  name: string;
  branch_id?: string;
}

export interface DepartmentUpdateRequest {
  code: string;
  name: string;
  branch_id?: string;
  is_active: boolean;
}

export const branchService = {
  list: (params: { page?: number; size?: number; q?: string; is_active?: boolean }) =>
    api.get<PaginatedResult<Branch>>('/branches', { params }).then(r => r.data),
  get: (id: string) => api.get<Branch>(`/branches/${id}`).then(r => r.data),
  create: (data: BranchCreateRequest) => api.post<Branch>('/branches', data).then(r => r.data),
  update: (id: string, data: BranchUpdateRequest) => api.put<Branch>(`/branches/${id}`, data).then(r => r.data),
};

export const departmentService = {
  list: (params: { page?: number; size?: number; q?: string; branch_id?: string; is_active?: boolean }) =>
    api.get<PaginatedResult<Department>>('/departments', { params }).then(r => r.data),
  get: (id: string) => api.get<Department>(`/departments/${id}`).then(r => r.data),
  create: (data: DepartmentCreateRequest) => api.post<Department>('/departments', data).then(r => r.data),
  update: (id: string, data: DepartmentUpdateRequest) => api.put<Department>(`/departments/${id}`, data).then(r => r.data),
};
