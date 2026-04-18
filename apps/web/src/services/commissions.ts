import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export type PlanType = 'flat' | 'tiered' | 'fixed' | 'custom';
export type RecordStatus = 'accrued' | 'approved' | 'paid' | 'clawback';
export type TriggerOn = 'invoice_issued' | 'payment_received' | 'engagement_completed';

export interface CommissionPlan {
  id: string;
  name: string;
  type: PlanType;
  trigger_on: TriggerOn;
  base_rate?: number;
  max_amount?: number;
  holdback_pct?: number;
  is_active: boolean;
  created_at: string;
}

export interface CommissionRecord {
  id: string;
  engagement_commission_id: string;
  salesperson_name?: string;
  engagement_title?: string;
  amount: number;
  status: RecordStatus;
  accrued_at: string;
  approved_at?: string;
  paid_at?: string;
  period_month?: string;
}

export interface CommissionSummary {
  ytd_accrued: number;
  ytd_paid: number;
  month_accrued: number;
  month_paid: number;
  pending_approval: number;
  on_hold: number;
}

export const commissionService = {
  plans: {
    list: (params?: { page?: number; size?: number }) =>
      api.get<PaginatedResult<CommissionPlan>>('/commission-plans', { params }).then(r => r.data),
    create: (data: Partial<CommissionPlan>) => api.post<CommissionPlan>('/commission-plans', data).then(r => r.data),
    deactivate: (id: string) => api.post(`/commission-plans/${id}/deactivate`).then(r => r.data),
  },
  records: {
    list: (params: { page?: number; size?: number; status?: string }) =>
      api.get<PaginatedResult<CommissionRecord>>('/commissions/records', { params }).then(r => r.data),
    approve: (id: string) => api.post(`/commissions/records/${id}/approve`).then(r => r.data),
    markPaid: (id: string) => api.post(`/commissions/records/${id}/mark-paid`).then(r => r.data),
    clawback: (id: string, reason: string) => api.post(`/commissions/records/${id}/clawback`, { reason }).then(r => r.data),
  },
  me: {
    list: (params: { page?: number; size?: number; status?: string; period?: string }) =>
      api.get<PaginatedResult<CommissionRecord>>('/me/commissions', { params }).then(r => r.data),
    summary: () => api.get<CommissionSummary>('/me/commissions/summary').then(r => r.data),
  },
};
