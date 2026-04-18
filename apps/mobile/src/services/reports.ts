import { api } from '@/lib/api';

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
  personal: () =>
    api.get<PersonalDashboard>('/reports/personal-dashboard').then(r => r.data),
};
