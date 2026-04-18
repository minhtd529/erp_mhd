import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export interface Employee {
  id: string;
  full_name: string;
  email: string;
  phone?: string;
  department?: string;
  position?: string;
  hire_date?: string;
  is_salesperson: boolean;
  sales_commission_eligible: boolean;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface EmployeeCreateRequest {
  full_name: string;
  email: string;
  phone?: string;
  department?: string;
  position?: string;
  hire_date?: string;
  is_salesperson?: boolean;
  sales_commission_eligible?: boolean;
}

export const employeeService = {
  list: (params: { page?: number; size?: number; q?: string }) =>
    api.get<PaginatedResult<Employee>>('/employees', { params }).then(r => r.data),
  get: (id: string) => api.get<Employee>(`/employees/${id}`).then(r => r.data),
  create: (data: EmployeeCreateRequest) => api.post<Employee>('/employees', data).then(r => r.data),
  update: (id: string, data: Partial<EmployeeCreateRequest>) => api.put<Employee>(`/employees/${id}`, data).then(r => r.data),
  delete: (id: string) => api.delete(`/employees/${id}`),
};
