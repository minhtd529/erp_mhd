import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export interface Client {
  id: string;
  business_name: string;
  english_name?: string;
  tax_code: string;
  address?: string;
  phone?: string;
  email?: string;
  representative_name?: string;
  representative_title?: string;
  sales_owner_id?: string;
  referrer_id?: string;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface ClientCreateRequest {
  business_name: string;
  english_name?: string;
  tax_code: string;
  address?: string;
  phone?: string;
  email?: string;
  representative_name?: string;
  representative_title?: string;
  sales_owner_id?: string;
  referrer_id?: string;
}

export const clientService = {
  list: (params: { page?: number; size?: number; q?: string }) =>
    api.get<PaginatedResult<Client>>('/clients', { params }).then(r => r.data),
  get: (id: string) => api.get<Client>(`/clients/${id}`).then(r => r.data),
  create: (data: ClientCreateRequest) => api.post<Client>('/clients', data).then(r => r.data),
  update: (id: string, data: Partial<ClientCreateRequest>) => api.put<Client>(`/clients/${id}`, data).then(r => r.data),
  delete: (id: string) => api.delete(`/clients/${id}`),
};
