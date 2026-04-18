import { api } from '@/lib/api';

export type EngagementStatus = 'DRAFT' | 'PROPOSAL' | 'CONTRACTED' | 'ACTIVE' | 'COMPLETED' | 'SETTLED';

export interface Engagement {
  id: string;
  code: string;
  client_id: string;
  client_name?: string;
  title: string;
  description?: string;
  status: EngagementStatus;
  service_type?: string;
  start_date?: string;
  end_date?: string;
  budget?: number;
  created_at: string;
  updated_at: string;
}

export interface PaginatedResult<T> {
  data: T[];
  total: number;
  page: number;
  size: number;
  total_pages: number;
}

export const STATUS_LABELS: Record<EngagementStatus, string> = {
  DRAFT: 'Nháp',
  PROPOSAL: 'Đề xuất',
  CONTRACTED: 'Đã ký HĐ',
  ACTIVE: 'Đang thực hiện',
  COMPLETED: 'Hoàn thành',
  SETTLED: 'Đã quyết toán',
};

export const engagementService = {
  list: (params: { page?: number; size?: number; q?: string; status?: string }) =>
    api.get<PaginatedResult<Engagement>>('/engagements', { params }).then(r => r.data),
  get: (id: string) =>
    api.get<Engagement>(`/engagements/${id}`).then(r => r.data),
};
