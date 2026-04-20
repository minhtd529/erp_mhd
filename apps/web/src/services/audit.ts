import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export interface AuditLogEntry {
  id: string;
  user_id?: string;
  user_name: string;
  module: string;
  resource: string;
  resource_id?: string;
  action: string;
  ip_address: string;
  created_at: string;
}

export interface AuditLogListParams {
  page?: number;
  size?: number;
  module?: string;
  resource?: string;
  action?: string;
  user_id?: string;
  from?: string;
  to?: string;
}

export const auditService = {
  list: (params: AuditLogListParams) =>
    api.get<PaginatedResult<AuditLogEntry>>('/audit-logs', { params }).then(r => r.data),
};

export const AUDIT_MODULES = [
  'global', 'org', 'hrm', 'crm', 'engagement',
  'timesheet', 'billing', 'workingpaper', 'commission', 'tax', 'reporting',
] as const;

export const AUDIT_ACTIONS = [
  'CREATE', 'UPDATE', 'DELETE',
  'UPDATE_BANK_DETAILS', 'UPDATE_STATUS', 'ASSIGN_ROLE',
  'APPROVE', 'REJECT', 'SUBMIT', 'LOCK', 'DEACTIVATE',
] as const;
