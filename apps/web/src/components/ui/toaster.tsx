'use client';
import { useToast } from '@/hooks/use-toast';
import { Toast, ToastClose, ToastTitle, ToastProvider, ToastViewport } from './toast';

export function Toaster() {
  const { toasts } = useToast();
  return (
    <ToastProvider>
      {toasts.map(({ id, title, variant, ...props }) => (
        <Toast key={id} variant={variant} {...props}>
          <ToastTitle>{title}</ToastTitle>
          <ToastClose />
        </Toast>
      ))}
      <ToastViewport />
    </ToastProvider>
  );
}
