'use client';
import * as React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { Department, UpdateDepartmentRequest } from '@/services/hrm/organization';

const schema = z.object({
  name: z.string().min(1, 'Bắt buộc'),
  description: z.string().optional(),
  dept_type: z.enum(['CORE', 'SUPPORT']),
});

export type DepartmentFormData = z.infer<typeof schema>;

interface DepartmentFormProps {
  formId: string;
  initial?: Partial<Department>;
  onSubmit: (data: UpdateDepartmentRequest) => void;
}

export function DepartmentForm({ formId, initial, onSubmit }: DepartmentFormProps) {
  const { register, handleSubmit, setValue, watch, formState: { errors } } = useForm<DepartmentFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initial?.name ?? '',
      description: initial?.description ?? '',
      dept_type: (initial?.dept_type as 'CORE' | 'SUPPORT') ?? 'CORE',
    },
  });

  const deptType = watch('dept_type');

  return (
    <form id={formId} onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="flex flex-col gap-1">
        <Label>Tên phòng ban *</Label>
        <Input {...register('name')} />
        {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
      </div>
      <div className="flex flex-col gap-1">
        <Label>Mô tả</Label>
        <Input {...register('description')} />
      </div>
      <div className="flex flex-col gap-1">
        <Label>Loại phòng ban *</Label>
        <Select value={deptType} onValueChange={(v) => setValue('dept_type', v as 'CORE' | 'SUPPORT')}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="CORE">CORE — Nghiệp vụ cốt lõi</SelectItem>
            <SelectItem value="SUPPORT">SUPPORT — Hỗ trợ</SelectItem>
          </SelectContent>
        </Select>
        {errors.dept_type && <p className="text-xs text-danger">{errors.dept_type.message}</p>}
      </div>
    </form>
  );
}
