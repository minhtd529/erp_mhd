import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

// ─── Types ────────────────────────────────────────────────────────────────────

export interface EmployeeListItem {
  id: string;
  employee_code?: string;
  full_name: string;
  display_name?: string;
  email: string;
  grade: string;
  status: string;
  branch_id?: string;
  department_id?: string;
  position_title?: string;
  employment_type: string;
  hired_date?: string;
  work_location: string;
  created_at: string;
}

export interface EmployeeDetail {
  id: string;
  employee_code?: string;
  full_name: string;
  display_name?: string;
  email: string;
  phone?: string;
  date_of_birth?: string;
  grade: string;
  status: string;
  manager_id?: string;
  branch_id?: string;
  department_id?: string;
  position_title?: string;
  employment_type: string;
  hired_date?: string;
  probation_end_date?: string;
  termination_date?: string;
  termination_reason?: string;
  current_contract_id?: string;
  gender?: string;
  place_of_birth?: string;
  nationality?: string;
  ethnicity?: string;
  personal_email?: string;
  personal_phone?: string;
  work_phone?: string;
  current_address?: string;
  permanent_address?: string;
  cccd_encrypted: string;
  cccd_issued_date?: string;
  cccd_issued_place?: string;
  passport_number?: string;
  passport_expiry?: string;
  hired_source?: string;
  probation_salary_pct?: number;
  work_location: string;
  remote_days_per_week?: number;
  education_level?: string;
  education_major?: string;
  education_school?: string;
  education_graduation_year?: number;
  vn_cpa_number?: string;
  vn_cpa_issued_date?: string;
  vn_cpa_expiry_date?: string;
  practicing_certificate_number?: string;
  practicing_certificate_expiry?: string;
  base_salary?: number;
  salary_currency?: string;
  salary_effective_date?: string;
  bank_account_encrypted: string;
  bank_name?: string;
  bank_branch?: string;
  mst_ca_nhan_encrypted: string;
  commission_rate?: number;
  commission_type: string;
  sales_target_yearly?: number;
  biz_dev_region?: string;
  so_bhxh_encrypted: string;
  bhxh_registered_date?: string;
  bhxh_province_code?: string;
  bhyt_card_number?: string;
  bhyt_expiry_date?: string;
  bhyt_registered_hospital_code?: string;
  bhyt_registered_hospital_name?: string;
  tncn_registered: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateEmployeeRequest {
  full_name: string;
  email: string;
  phone?: string;
  date_of_birth?: string;
  grade: string;
  manager_id?: string;
  status?: string;
  branch_id?: string;
  department_id?: string;
  position_title?: string;
  employment_type?: string;
  hired_date?: string;
  display_name?: string;
  gender?: string;
  personal_email?: string;
  personal_phone?: string;
  work_location?: string;
  hired_source?: string;
  education_level?: string;
  commission_type?: string;
}

export interface UpdateEmployeeRequest {
  full_name?: string;
  phone?: string;
  grade?: string;
  manager_id?: string;
  status?: string;
  branch_id?: string;
  department_id?: string;
  position_title?: string;
  employment_type?: string;
  hired_date?: string;
  probation_end_date?: string;
  termination_date?: string;
  termination_reason?: string;
  display_name?: string;
  gender?: string;
  personal_email?: string;
  personal_phone?: string;
  work_phone?: string;
  current_address?: string;
  permanent_address?: string;
  work_location?: string;
  remote_days_per_week?: number;
  hired_source?: string;
  education_level?: string;
  education_major?: string;
  education_school?: string;
  education_graduation_year?: number;
  vn_cpa_number?: string;
  practicing_certificate_number?: string;
  commission_type?: string;
  commission_rate?: number;
  biz_dev_region?: string;
  nationality?: string;
  ethnicity?: string;
}

export interface UpdateProfileRequest {
  display_name?: string;
  personal_phone?: string;
  personal_email?: string;
  current_address?: string;
  permanent_address?: string;
}

export interface SensitiveFields {
  id: string;
  employee_code: string;
  full_name: string;
  cccd?: string;
  cccd_issued_date?: string;
  cccd_issued_place?: string;
  passport_number?: string;
  passport_expiry?: string;
  mst_ca_nhan?: string;
  so_bhxh?: string;
  bank_account?: string;
  bank_name?: string;
  bank_branch?: string;
  accessed_at: string;
}

export interface UpdateSensitiveRequest {
  cccd?: string;
  cccd_issued_date?: string;
  cccd_issued_place?: string;
  passport_number?: string;
  passport_expiry?: string;
  mst_ca_nhan?: string;
  so_bhxh?: string;
  bank_account?: string;
  bank_name?: string;
  bank_branch?: string;
}

export interface Dependent {
  id: string;
  employee_id: string;
  full_name: string;
  relationship: string;
  date_of_birth?: string;
  cccd_or_birth_cert?: string;
  tax_deduction_registered: boolean;
  tax_deduction_from?: string;
  tax_deduction_to?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDependentRequest {
  full_name: string;
  relationship: string;
  date_of_birth?: string;
  cccd_or_birth_cert?: string;
  tax_deduction_registered?: boolean;
  tax_deduction_from?: string;
  tax_deduction_to?: string;
  notes?: string;
}

export interface Contract {
  id: string;
  employee_id: string;
  contract_number?: string;
  contract_type: string;
  start_date: string;
  end_date?: string;
  signed_date?: string;
  salary_at_signing?: number;
  position_at_signing?: string;
  notes?: string;
  document_url?: string;
  is_current: boolean;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateContractRequest {
  contract_number?: string;
  contract_type: string;
  start_date: string;
  end_date?: string;
  signed_date?: string;
  salary_at_signing?: number;
  position_at_signing?: string;
  notes?: string;
  document_url?: string;
}

export interface SalaryHistoryItem {
  id: string;
  employee_id: string;
  effective_from: string;
  base_salary: string;
  allowances_total?: string;
  change_type: string;
  reason?: string;
  created_by_name?: string;
  created_at: string;
}

export interface CreateSalaryHistoryRequest {
  effective_from: string;
  base_salary: string;
  allowances_total?: string;
  change_type: string;
  reason?: string;
}

// ─── Employee service ─────────────────────────────────────────────────────────

export const employeeService = {
  list: (params?: { page?: number; size?: number; q?: string; status?: string; grade?: string; branch_id?: string; department_id?: string }) =>
    api.get<PaginatedResult<EmployeeListItem>>('/hrm/employees', { params }).then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: EmployeeDetail }>(`/hrm/employees/${id}`).then(r => r.data.data),

  create: (data: CreateEmployeeRequest) =>
    api.post<{ data: EmployeeDetail }>('/hrm/employees', data).then(r => r.data.data),

  update: (id: string, data: UpdateEmployeeRequest) =>
    api.put<{ data: EmployeeDetail }>(`/hrm/employees/${id}`, data).then(r => r.data.data),

  delete: (id: string) =>
    api.delete(`/hrm/employees/${id}`).then(r => r.data),

  // Dependents
  listDependents: (id: string) =>
    api.get<{ data: Dependent[] }>(`/hrm/employees/${id}/dependents`).then(r => r.data.data),

  createDependent: (id: string, data: CreateDependentRequest) =>
    api.post<{ data: Dependent }>(`/hrm/employees/${id}/dependents`, data).then(r => r.data.data),

  updateDependent: (id: string, depId: string, data: Partial<CreateDependentRequest>) =>
    api.put<{ data: Dependent }>(`/hrm/employees/${id}/dependents/${depId}`, data).then(r => r.data.data),

  deleteDependent: (id: string, depId: string) =>
    api.delete(`/hrm/employees/${id}/dependents/${depId}`).then(r => r.data),

  // Contracts
  listContracts: (id: string) =>
    api.get<{ data: Contract[] }>(`/hrm/employees/${id}/contracts`).then(r => r.data.data),

  createContract: (id: string, data: CreateContractRequest) =>
    api.post<{ data: Contract }>(`/hrm/employees/${id}/contracts`, data).then(r => r.data.data),

  updateContract: (id: string, cid: string, data: Partial<CreateContractRequest>) =>
    api.put<{ data: Contract }>(`/hrm/employees/${id}/contracts/${cid}`, data).then(r => r.data.data),

  terminateContract: (id: string, cid: string) =>
    api.post(`/hrm/employees/${id}/contracts/${cid}/terminate`).then(r => r.data),

  // Sensitive PII
  getSensitive: (id: string) =>
    api.get<{ data: SensitiveFields }>(`/hrm/employees/${id}/sensitive`).then(r => r.data.data),

  updateSensitive: (id: string, data: UpdateSensitiveRequest) =>
    api.put(`/hrm/employees/${id}/sensitive`, data).then(r => r.data),

  // Salary history
  listSalaryHistory: (id: string) =>
    api.get<{ data: SalaryHistoryItem[] }>(`/hrm/employees/${id}/salary-history`).then(r => r.data.data),

  createSalaryHistory: (id: string, data: CreateSalaryHistoryRequest) =>
    api.post<{ data: SalaryHistoryItem }>(`/hrm/employees/${id}/salary-history`, data).then(r => r.data.data),
};

// ─── My profile service ───────────────────────────────────────────────────────

export const profileService = {
  get: () =>
    api.get<{ data: EmployeeDetail }>('/me/hrm-profile').then(r => r.data.data),

  update: (data: UpdateProfileRequest) =>
    api.put<{ data: EmployeeDetail }>('/me/hrm-profile', data).then(r => r.data.data),
};
