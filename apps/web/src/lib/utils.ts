import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatCurrency(amount: number, currency = 'VND') {
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency }).format(amount);
}

export function formatDate(date: string | Date) {
  return new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short' }).format(new Date(date));
}

export function formatDateTime(date: string | Date) {
  return new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(date));
}

export function getErrorMessage(err: unknown): string {
  if (err && typeof err === 'object' && 'response' in err) {
    const r = (err as { response?: { data?: { message?: string } } }).response;
    return r?.data?.message ?? 'Đã xảy ra lỗi';
  }
  return 'Đã xảy ra lỗi';
}
