import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface ProvisioningRequest {
  id: string;
  employee_id: string;
  requested_by: string;
  requested_role: string;
  requested_branch_id?: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'EXECUTED' | 'CANCELLED';
  approval_level: number;
  branch_approver_id?: string;
  branch_approved_at?: string;
  branch_rejection_reason?: string;
  hr_approver_id?: string;
  hr_approved_at?: string;
  hr_rejection_reason?: string;
  executed_by?: string;
  executed_at?: string;
  is_emergency: boolean;
  emergency_reason?: string;
  notes?: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
}

export interface ExecuteProvisioningResponse {
  request: ProvisioningRequest;
  user_id: string;
  temp_password: string;
}

export interface OffboardingChecklist {
  id: string;
  employee_id: string;
  checklist_type: 'ONBOARDING' | 'OFFBOARDING';
  initiated_by: string;
  target_date?: string;
  items: OffboardingItems;
  status: 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED';
  completed_at?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface OffboardingItem {
  key: string;
  label: string;
  completed: boolean;
  notes?: string;
}

export interface OffboardingItems {
  items: OffboardingItem[];
}

// ─── Request types ─────────────────────────────────────────────────────────────

export interface CreateProvisioningRequest {
  employee_id: string;
  requested_role: string;
  requested_branch_id?: string;
  is_emergency?: boolean;
  emergency_reason?: string;
  notes?: string;
}

export interface RejectRequest {
  reason: string;
}

export interface ExecuteRequest {
  full_name: string;
  email: string;
}

export interface CreateOffboardingRequest {
  employee_id: string;
  checklist_type: 'ONBOARDING' | 'OFFBOARDING';
  target_date?: string;
  notes?: string;
}

export interface UpdateOffboardingItemRequest {
  completed: boolean;
  notes?: string;
}

export interface ListProvisioningParams {
  page?: number;
  size?: number;
  status?: string;
  employee_id?: string;
}

// ─── Services ─────────────────────────────────────────────────────────────────

export const provisioningService = {
  list: (params?: ListProvisioningParams) =>
    api.get<PaginatedResult<ProvisioningRequest>>('/hrm/user-provisioning-requests', { params })
       .then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}`)
       .then(r => r.data.data),

  create: (data: CreateProvisioningRequest) =>
    api.post<{ data: ProvisioningRequest }>('/hrm/user-provisioning-requests', data)
       .then(r => r.data.data),

  branchApprove: (id: string) =>
    api.post<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}/branch-approve`)
       .then(r => r.data.data),

  branchReject: (id: string, data: RejectRequest) =>
    api.post<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}/branch-reject`, data)
       .then(r => r.data.data),

  hrApprove: (id: string) =>
    api.post<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}/hr-approve`)
       .then(r => r.data.data),

  hrReject: (id: string, data: RejectRequest) =>
    api.post<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}/hr-reject`, data)
       .then(r => r.data.data),

  execute: (id: string, data: ExecuteRequest) =>
    api.post<{ data: ExecuteProvisioningResponse }>(`/hrm/user-provisioning-requests/${id}/execute`, data)
       .then(r => r.data.data),

  cancel: (id: string) =>
    api.post<{ data: ProvisioningRequest }>(`/hrm/user-provisioning-requests/${id}/cancel`)
       .then(r => r.data.data),
};

export const offboardingService = {
  list: (params?: { page?: number; size?: number; status?: string; employee_id?: string }) =>
    api.get<PaginatedResult<OffboardingChecklist>>('/hrm/offboarding', { params })
       .then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: OffboardingChecklist }>(`/hrm/offboarding/${id}`)
       .then(r => r.data.data),

  create: (data: CreateOffboardingRequest) =>
    api.post<{ data: OffboardingChecklist }>('/hrm/offboarding', data)
       .then(r => r.data.data),

  updateItem: (id: string, key: string, data: UpdateOffboardingItemRequest) =>
    api.put<{ data: OffboardingChecklist }>(`/hrm/offboarding/${id}/items/${key}`, data)
       .then(r => r.data.data),

  complete: (id: string) =>
    api.post<{ data: OffboardingChecklist }>(`/hrm/offboarding/${id}/complete`)
       .then(r => r.data.data),
};
