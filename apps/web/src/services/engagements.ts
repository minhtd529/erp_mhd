import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

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
  primary_salesperson_id?: string;
  created_at: string;
  updated_at: string;
}

export interface EngagementCreateRequest {
  client_id: string;
  title: string;
  description?: string;
  service_type?: string;
  start_date?: string;
  end_date?: string;
  budget?: number;
}

const STATUS_TRANSITIONS: Record<EngagementStatus, string[]> = {
  DRAFT: ['submit'],
  PROPOSAL: ['contract', 'reject'],
  CONTRACTED: ['activate'],
  ACTIVE: ['complete'],
  COMPLETED: ['settle'],
  SETTLED: [],
};

export const engagementService = {
  list: (params: { page?: number; size?: number; q?: string; status?: string }) =>
    api.get<PaginatedResult<Engagement>>('/engagements', { params }).then(r => r.data),
  get: (id: string) => api.get<Engagement>(`/engagements/${id}`).then(r => r.data),
  create: (data: EngagementCreateRequest) => api.post<Engagement>('/engagements', data).then(r => r.data),
  update: (id: string, data: Partial<EngagementCreateRequest>) => api.put<Engagement>(`/engagements/${id}`, data).then(r => r.data),
  delete: (id: string) => api.delete(`/engagements/${id}`),
  transition: (id: string, action: string) => api.post(`/engagements/${id}/${action}`).then(r => r.data),
  getAvailableTransitions: (status: EngagementStatus) => STATUS_TRANSITIONS[status] ?? [],
};
