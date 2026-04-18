import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export type TimesheetStatus = 'OPEN' | 'SUBMITTED' | 'APPROVED' | 'REJECTED' | 'LOCKED';

export interface Timesheet {
  id: string;
  employee_id: string;
  employee_name?: string;
  period_start: string;
  period_end: string;
  status: TimesheetStatus;
  total_hours?: number;
  submitted_at?: string;
  approved_at?: string;
  created_at: string;
}

export interface TimesheetEntry {
  id: string;
  timesheet_id: string;
  engagement_id: string;
  engagement_title?: string;
  date: string;
  hours: number;
  description?: string;
}

export const timesheetService = {
  list: (params: { page?: number; size?: number; employee_id?: string; status?: string }) =>
    api.get<PaginatedResult<Timesheet>>('/timesheets', { params }).then(r => r.data),
  get: (id: string) => api.get<Timesheet>(`/timesheets/${id}`).then(r => r.data),
  submit: (id: string) => api.post(`/timesheets/${id}/submit`).then(r => r.data),
  approve: (id: string) => api.post(`/timesheets/${id}/approve`).then(r => r.data),
  reject: (id: string, reason?: string) => api.post(`/timesheets/${id}/reject`, { reason }).then(r => r.data),
  lock: (id: string) => api.post(`/timesheets/${id}/lock`).then(r => r.data),
  listEntries: (timesheetId: string) =>
    api.get<PaginatedResult<TimesheetEntry>>(`/timesheets/${timesheetId}/entries`).then(r => r.data),
  createEntry: (timesheetId: string, data: Omit<TimesheetEntry, 'id' | 'timesheet_id'>) =>
    api.post(`/timesheets/${timesheetId}/entries`, data).then(r => r.data),
  deleteEntry: (timesheetId: string, entryId: string) =>
    api.delete(`/timesheets/${timesheetId}/entries/${entryId}`),
};
