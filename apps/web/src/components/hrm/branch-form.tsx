'use client';
import * as React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import type { Branch, UpdateBranchRequest } from '@/services/hrm/organization';

const schema = z.object({
  name: z.string().min(1, 'Bắt buộc').optional(),
  address: z.string().optional(),
  phone: z.string().optional(),
  city: z.string().optional(),
  tax_code: z.string().optional(),
  established_date: z.string().optional(),
  authorization_doc_number: z.string().optional(),
  authorization_date: z.string().optional(),
});

export type BranchFormData = z.infer<typeof schema>;

interface BranchFormProps {
  formId: string;
  initial?: Partial<Branch>;
  canEditCritical: boolean;
  onSubmit: (data: UpdateBranchRequest) => void;
}

export function BranchForm({ formId, initial, canEditCritical, onSubmit }: BranchFormProps) {
  const { register, handleSubmit, formState: { errors } } = useForm<BranchFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initial?.name ?? '',
      address: initial?.address ?? '',
      phone: initial?.phone ?? '',
      city: initial?.city ?? '',
      tax_code: initial?.tax_code ?? '',
      established_date: initial?.established_date ?? '',
      authorization_doc_number: initial?.authorization_doc_number ?? '',
      authorization_date: initial?.authorization_date ?? '',
    },
  });

  return (
    <form id={formId} onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        {canEditCritical && (
          <div className="flex flex-col gap-1 col-span-2">
            <Label>Tên chi nhánh *</Label>
            <Input {...register('name')} />
            {errors.name && <p className="text-xs text-danger">{errors.name.message}</p>}
          </div>
        )}
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Địa chỉ</Label>
          <Input {...register('address')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Thành phố</Label>
          <Input {...register('city')} />
        </div>
        {canEditCritical && (
          <>
            <div className="flex flex-col gap-1">
              <Label>Mã số thuế</Label>
              <Input {...register('tax_code')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ngày thành lập</Label>
              <Input {...register('established_date')} type="date" />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số quyết định ủy quyền</Label>
              <Input {...register('authorization_doc_number')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ngày ủy quyền</Label>
              <Input {...register('authorization_date')} type="date" />
            </div>
          </>
        )}
      </div>
    </form>
  );
}
