'use client';
import * as React from 'react';
import type { ToastVariant } from '@/components/ui/toast';

interface ToastItem {
  id: string;
  title: string;
  variant?: ToastVariant;
}

let toastListeners: Array<(toasts: ToastItem[]) => void> = [];
let toastList: ToastItem[] = [];

function dispatch(toasts: ToastItem[]) {
  toastList = toasts;
  toastListeners.forEach((l) => l(toasts));
}

export function toast(title: string, variant: ToastVariant = 'default') {
  const id = Math.random().toString(36).slice(2);
  dispatch([...toastList, { id, title, variant }]);
  setTimeout(() => {
    dispatch(toastList.filter((t) => t.id !== id));
  }, 5000);
}

export function useToast() {
  const [toasts, setToasts] = React.useState<ToastItem[]>(toastList);
  React.useEffect(() => {
    toastListeners.push(setToasts);
    return () => {
      toastListeners = toastListeners.filter((l) => l !== setToasts);
    };
  }, []);
  return { toasts };
}
