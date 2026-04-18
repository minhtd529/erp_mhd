import { api } from '@/lib/api';
import type { PaginatedResult } from './engagements';

export type TimesheetStatus = 'OPEN' | 'SUBMITTED' | 'APPROVED' | 'REJECTED' | 'LOCKED';

export interface Timesheet {
  id: string;
  employee_id: string;
  period_start: string;
  period_end: string;
  status: TimesheetStatus;
  total_hours?: number;
  submitted_at?: string;
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

export const STATUS_LABELS: Record<TimesheetStatus, string> = {
  OPEN: 'Đang mở',
  SUBMITTED: 'Chờ duyệt',
  APPROVED: 'Đã duyệt',
  REJECTED: 'Từ chối',
  LOCKED: 'Đã khóa',
};

export const timesheetService = {
  list: (params?: { page?: number; size?: number }) =>
    api.get<PaginatedResult<Timesheet>>('/timesheets', { params }).then(r => r.data),
  get: (id: string) =>
    api.get<Timesheet>(`/timesheets/${id}`).then(r => r.data),
  submit: (id: string) =>
    api.post(`/timesheets/${id}/submit`).then(r => r.data),
  listEntries: (timesheetId: string) =>
    api.get<PaginatedResult<TimesheetEntry>>(`/timesheets/${timesheetId}/entries`).then(r => r.data),
  createEntry: (timesheetId: string, data: { engagement_id: string; date: string; hours: number; description?: string }) =>
    api.post<TimesheetEntry>(`/timesheets/${timesheetId}/entries`, data).then(r => r.data),
  deleteEntry: (timesheetId: string, entryId: string) =>
    api.delete(`/timesheets/${timesheetId}/entries/${entryId}`),
};
