import { api } from '@/lib/api';
import type { PaginatedResult } from '@/types';

export type InvoiceStatus = 'DRAFT' | 'SENT' | 'CONFIRMED' | 'ISSUED' | 'PAID' | 'CANCELLED';
export type PaymentStatus = 'RECORDED' | 'CLEARED' | 'DISPUTED' | 'REVERSED';

export interface Invoice {
  id: string;
  number: string;
  client_id: string;
  client_name?: string;
  engagement_id?: string;
  status: InvoiceStatus;
  total_amount: number;
  tax_amount?: number;
  issued_date?: string;
  due_date?: string;
  created_at: string;
}

export interface Payment {
  id: string;
  invoice_id: string;
  amount: number;
  payment_method?: string;
  status: PaymentStatus;
  payment_date?: string;
  reference_number?: string;
  created_at: string;
}

export const invoiceService = {
  list: (params: { page?: number; size?: number; status?: string; client_id?: string }) =>
    api.get<PaginatedResult<Invoice>>('/invoices', { params }).then(r => r.data),
  get: (id: string) => api.get<Invoice>(`/invoices/${id}`).then(r => r.data),
  create: (data: Partial<Invoice>) => api.post<Invoice>('/invoices', data).then(r => r.data),
  update: (id: string, data: Partial<Invoice>) => api.put<Invoice>(`/invoices/${id}`, data).then(r => r.data),
  transition: (id: string, action: string) => api.post(`/invoices/${id}/${action}`).then(r => r.data),
  approvalQueue: () => api.get<PaginatedResult<Invoice>>('/invoices/approval-queue').then(r => r.data),
  exportCSV: () => api.get('/invoices/export', { responseType: 'blob' }).then(r => r.data),
};

export const paymentService = {
  list: (params: { page?: number; invoice_id?: string }) =>
    api.get<PaginatedResult<Payment>>('/payments', { params }).then(r => r.data),
  record: (data: { invoice_id: string; amount: number; payment_method?: string; reference_number?: string }) =>
    api.post<Payment>('/payments', data).then(r => r.data),
  clear: (id: string) => api.post(`/payments/${id}/clear`).then(r => r.data),
  dispute: (id: string) => api.post(`/payments/${id}/dispute`).then(r => r.data),
  reverse: (id: string) => api.delete(`/payments/${id}`).then(r => r.data),
};
