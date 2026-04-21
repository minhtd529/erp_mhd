'use client';
import * as React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';

const schema = z.object({
  id: z.string().uuid('ID không hợp lệ'),
});

type FormData = z.infer<typeof schema>;

interface AssignHeadDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  label: string;
  placeholder?: string;
  loading?: boolean;
  onSubmit: (id: string) => void;
}

export function AssignHeadDialog({ open, onOpenChange, title, label, placeholder, loading, onSubmit }: AssignHeadDialogProps) {
  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  React.useEffect(() => {
    if (!open) reset();
  }, [open, reset]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <form id="assign-head-form" onSubmit={handleSubmit((d) => onSubmit(d.id))} className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <Label>{label}</Label>
            <Input {...register('id')} placeholder={placeholder ?? 'Nhập UUID...'} />
            {errors.id && <p className="text-xs text-danger">{errors.id.message}</p>}
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Hủy</Button>
          <Button type="submit" form="assign-head-form" loading={loading}>Xác nhận</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
