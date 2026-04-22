import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

// ─── Types ────────────────────────────────────────────────────────────────────

export type CertType =
  | 'VN_CPA' | 'ACCA' | 'CPA_AUSTRALIA' | 'CFA' | 'CIA'
  | 'CISA' | 'IFRS' | 'ICAEW' | 'CMA' | 'OTHER';

export type CertStatus = 'ACTIVE' | 'EXPIRED' | 'REVOKED' | 'SUSPENDED';

export type CourseType =
  | 'TECHNICAL' | 'ETHICS' | 'MANAGEMENT' | 'SOFT_SKILLS' | 'COMPLIANCE' | 'OTHER';

export type TrainingStatus =
  | 'ENROLLED' | 'IN_PROGRESS' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

export interface Certification {
  id: string;
  employee_id: string;
  cert_type: CertType;
  cert_number: string;
  issued_date: string;
  expiry_date?: string;
  issuing_authority: string;
  status: CertStatus;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface TrainingCourse {
  id: string;
  code: string;
  name: string;
  description?: string;
  course_type: CourseType;
  cpe_hours: number;
  provider?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TrainingRecord {
  id: string;
  employee_id: string;
  course_id: string;
  course_name?: string;
  course_type?: CourseType;
  completion_date?: string;
  status: TrainingStatus;
  cpe_hours_earned: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CPERequirement {
  id: string;
  role_code: string;
  year: number;
  required_hours: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CPESummary {
  employee_id: string;
  year: number;
  total_hours_earned: number;
  required_hours: number;
  by_category: Record<string, number>;
}

// ─── Request types ─────────────────────────────────────────────────────────────

export interface CreateCertificationRequest {
  cert_type: CertType;
  cert_name: string;
  cert_number: string;
  issued_date: string;
  expiry_date?: string;
  issuing_authority: string;
  status?: CertStatus;
  notes?: string;
}

export interface UpdateCertificationRequest {
  cert_type?: CertType;
  cert_name?: string;
  cert_number?: string;
  issued_date?: string;
  expiry_date?: string;
  issuing_authority?: string;
  status?: CertStatus;
  notes?: string;
}

export interface CreateTrainingCourseRequest {
  code: string;
  name: string;
  description?: string;
  course_type: CourseType;
  cpe_hours: number;
  provider?: string;
  is_active?: boolean;
}

export interface UpdateTrainingCourseRequest {
  name?: string;
  description?: string;
  course_type?: CourseType;
  cpe_hours?: number;
  provider?: string;
  is_active?: boolean;
}

export interface CreateTrainingRecordRequest {
  course_id: string;
  status?: TrainingStatus;
  cpe_hours_earned?: number;
  notes?: string;
}

export interface UpdateTrainingRecordRequest {
  completion_date?: string;
  status?: TrainingStatus;
  cpe_hours_earned?: number;
  notes?: string;
}

export interface CreateCPERequirementRequest {
  role_code: string;
  year: number;
  required_hours: number;
  notes?: string;
}

export interface UpdateCPERequirementRequest {
  required_hours?: number;
  notes?: string;
}

export interface ListCertificationsParams {
  page?: number;
  size?: number;
  status?: CertStatus;
  cert_type?: CertType;
}

export interface ListTrainingCoursesParams {
  page?: number;
  size?: number;
  course_type?: CourseType;
  is_active?: boolean;
}

export interface ListTrainingRecordsParams {
  page?: number;
  size?: number;
  status?: TrainingStatus;
}

export interface ListCPERequirementsParams {
  page?: number;
  size?: number;
  year?: number;
  role_code?: string;
}

// ─── Services ─────────────────────────────────────────────────────────────────

export const certificationService = {
  listByEmployee: (employeeId: string, params?: ListCertificationsParams) =>
    api.get<PaginatedResult<Certification>>(`/hrm/employees/${employeeId}/certifications`, { params })
       .then(r => r.data),

  create: (employeeId: string, data: CreateCertificationRequest) =>
    api.post<{ data: Certification }>(`/hrm/employees/${employeeId}/certifications`, data)
       .then(r => r.data.data),

  getById: (id: string) =>
    api.get<{ data: Certification }>(`/hrm/certifications/${id}`)
       .then(r => r.data.data),

  update: (id: string, data: UpdateCertificationRequest) =>
    api.put<{ data: Certification }>(`/hrm/certifications/${id}`, data)
       .then(r => r.data.data),

  delete: (id: string) =>
    api.delete(`/hrm/certifications/${id}`),

  listExpiring: (params?: { days?: number; page?: number; size?: number }) =>
    api.get<PaginatedResult<Certification>>('/hrm/certifications/expiring', { params })
       .then(r => r.data),
};

export const trainingCourseService = {
  list: (params?: ListTrainingCoursesParams) =>
    api.get<PaginatedResult<TrainingCourse>>('/hrm/training-courses', { params })
       .then(r => r.data),

  getById: (id: string) =>
    api.get<{ data: TrainingCourse }>(`/hrm/training-courses/${id}`)
       .then(r => r.data.data),

  create: (data: CreateTrainingCourseRequest) =>
    api.post<{ data: TrainingCourse }>('/hrm/training-courses', data)
       .then(r => r.data.data),

  update: (id: string, data: UpdateTrainingCourseRequest) =>
    api.put<{ data: TrainingCourse }>(`/hrm/training-courses/${id}`, data)
       .then(r => r.data.data),

  delete: (id: string) =>
    api.delete(`/hrm/training-courses/${id}`),
};

export const trainingRecordService = {
  listByEmployee: (employeeId: string, params?: ListTrainingRecordsParams) =>
    api.get<PaginatedResult<TrainingRecord>>(`/hrm/employees/${employeeId}/training-records`, { params })
       .then(r => r.data),

  create: (employeeId: string, data: CreateTrainingRecordRequest) =>
    api.post<{ data: TrainingRecord }>(`/hrm/employees/${employeeId}/training-records`, data)
       .then(r => r.data.data),

  getById: (id: string) =>
    api.get<{ data: TrainingRecord }>(`/hrm/training-records/${id}`)
       .then(r => r.data.data),

  update: (id: string, data: UpdateTrainingRecordRequest) =>
    api.put<{ data: TrainingRecord }>(`/hrm/training-records/${id}`, data)
       .then(r => r.data.data),

  delete: (id: string) =>
    api.delete(`/hrm/training-records/${id}`),
};

export const cpeRequirementService = {
  list: (params?: ListCPERequirementsParams) =>
    api.get<PaginatedResult<CPERequirement>>('/hrm/cpe-requirements', { params })
       .then(r => r.data),

  create: (data: CreateCPERequirementRequest) =>
    api.post<{ data: CPERequirement }>('/hrm/cpe-requirements', data)
       .then(r => r.data.data),

  update: (id: string, data: UpdateCPERequirementRequest) =>
    api.put<{ data: CPERequirement }>(`/hrm/cpe-requirements/${id}`, data)
       .then(r => r.data.data),

  getSummary: (employeeId: string, year: number) =>
    api.get<{ data: CPESummary }>(`/hrm/employees/${employeeId}/cpe-summary`, { params: { year } })
       .then(r => r.data.data),
};
