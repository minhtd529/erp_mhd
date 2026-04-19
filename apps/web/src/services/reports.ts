import { api } from '@/lib/api';

export interface ExecutiveDashboard {
  total_revenue: number;
  total_clients: number;
  active_engagements: number;
  outstanding_ar: number;
  commission_kpis: {
    accrued: number;
    paid: number;
    pending: number;
    on_hold: number;
    commission_pct_of_revenue: number;
  };
}

export interface ManagerDashboard {
  team_active_engagements: number;
  team_pending_timesheets: number;
  team_utilization_rate: number;
  top_engagements: Array<{ id: string; title: string; budget: number; progress_pct: number }>;
}

export interface PersonalDashboard {
  active_engagements: number;
  pending_timesheets: number;
  total_hours_this_month: number;
  is_salesperson: boolean;
  commission_ytd?: number;
  commission_month?: number;
  commission_pending?: number;
  commission_on_hold?: number;
}

export const reportService = {
  executive: () => api.get<ExecutiveDashboard>('/dashboard/executive').then(r => r.data),
  manager: () => api.get<ManagerDashboard>('/dashboard/manager').then(r => r.data),
  personal: () => api.get<PersonalDashboard>('/dashboard/personal').then(r => r.data),
  periodSummary: (start: string, end: string) =>
    api.get('/billing/reports/period-summary', { params: { start, end } }).then(r => r.data),
  paymentSummary: (start: string, end: string) =>
    api.get('/billing/reports/payment-summary', { params: { start, end } }).then(r => r.data),
  commissionPayout: (params?: { months?: number }) =>
    api.get('/reports/commission-payout', { params }).then(r => r.data),
  revenueByStaff: () => api.get('/reports/revenue-by-salesperson').then(r => r.data),
};
