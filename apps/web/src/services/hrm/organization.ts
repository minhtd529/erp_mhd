import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Branch {
  id: string;
  code: string;
  name: string;
  address?: string;
  phone?: string;
  city?: string;
  tax_code?: string;
  is_head_office: boolean;
  is_active: boolean;
  established_date?: string;
  head_of_branch_user_id?: string;
  authorization_doc_number?: string;
  authorization_date?: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateBranchRequest {
  address?: string;
  phone?: string;
  city?: string;
  name?: string;
  tax_code?: string;
  established_date?: string;
  authorization_doc_number?: string;
  authorization_date?: string;
}

export interface Department {
  id: string;
  code: string;
  name: string;
  description?: string;
  dept_type: string;
  head_employee_id?: string;
  is_active: boolean;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpdateDepartmentRequest {
  name?: string;
  description?: string;
  dept_type?: string;
  authorization_doc_number?: string;
  authorization_date?: string;
}

export interface BranchDepartment {
  branch_id: string;
  department_id: string;
  head_employee_id?: string;
  is_active: boolean;
  created_at: string;
}

export interface OrgChartBranch {
  id: string;
  code: string;
  name: string;
  is_head_office: boolean;
  departments: OrgChartDept[];
}

export interface OrgChartDept {
  id: string;
  code: string;
  name: string;
  dept_type: string;
}

export interface OrgChartResponse {
  branches: OrgChartBranch[];
}

// ─── Branch service ───────────────────────────────────────────────────────────

export const branchService = {
  list: (params?: { page?: number; size?: number; q?: string; is_active?: boolean }) =>
    api.get<PaginatedResult<Branch>>('/hrm/organization/branches', { params }).then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: Branch }>(`/hrm/organization/branches/${id}`).then(r => r.data.data),

  update: (id: string, data: UpdateBranchRequest) =>
    api.put<{ data: Branch }>(`/hrm/organization/branches/${id}`, data).then(r => r.data.data),

  assignHead: (id: string, userId: string) =>
    api.put<{ data: Branch }>(`/hrm/organization/branches/${id}/assign-head`, { user_id: userId }).then(r => r.data.data),

  deactivate: (id: string) =>
    api.put(`/hrm/organization/branches/${id}/deactivate`).then(r => r.data),
};

// ─── Department service ───────────────────────────────────────────────────────

export const departmentService = {
  list: (params?: { page?: number; size?: number; q?: string; is_active?: boolean }) =>
    api.get<PaginatedResult<Department>>('/hrm/organization/departments', { params }).then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: Department }>(`/hrm/organization/departments/${id}`).then(r => r.data.data),

  update: (id: string, data: UpdateDepartmentRequest) =>
    api.put<{ data: Department }>(`/hrm/organization/departments/${id}`, data).then(r => r.data.data),

  assignHead: (id: string, employeeId: string) =>
    api.put<{ data: Department }>(`/hrm/organization/departments/${id}/assign-head`, { employee_id: employeeId }).then(r => r.data.data),

  deactivate: (id: string) =>
    api.put(`/hrm/organization/departments/${id}/deactivate`).then(r => r.data),
};

// ─── Matrix service ───────────────────────────────────────────────────────────

export const matrixService = {
  list: () =>
    api.get<PaginatedResult<BranchDepartment>>('/hrm/organization/branch-departments').then(r => r.data),

  link: (branchId: string, departmentId: string) =>
    api.post('/hrm/organization/branch-departments', { branch_id: branchId, department_id: departmentId }).then(r => r.data),

  unlink: (branchId: string, deptId: string) =>
    api.delete(`/hrm/organization/branch-departments/${branchId}/${deptId}`).then(r => r.data),
};

// ─── Org chart service ────────────────────────────────────────────────────────

export const orgChartService = {
  get: () =>
    api.get<{ data: OrgChartResponse }>('/hrm/organization/org-chart').then(r => r.data.data),
};
