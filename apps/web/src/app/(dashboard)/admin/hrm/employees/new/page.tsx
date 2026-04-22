'use client';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { CreateEmployeeForm } from '@/components/hrm/employee-form';
import { employeeService, type CreateEmployeeRequest } from '@/services/hrm/employee';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { ArrowLeft } from 'lucide-react';

export default function NewEmployeePage() {
  const router = useRouter();

  const createMutation = useMutation({
    mutationFn: (data: CreateEmployeeRequest) => employeeService.create(data),
    onSuccess: (emp) => {
      toast('Tạo nhân viên thành công', 'success');
      router.push(`/admin/hrm/employees/${emp.id}`);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-4 max-w-3xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/employees')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Thêm nhân viên mới</h1>
          <p className="text-sm text-text-secondary">Điền thông tin để tạo hồ sơ nhân viên</p>
        </div>
      </div>

      <Card>
        <CardContent className="p-5">
          <CreateEmployeeForm formId="create-employee-form" onSubmit={(data) => createMutation.mutate(data)} />
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={() => router.push('/admin/hrm/employees')}>Hủy</Button>
        <Button type="submit" form="create-employee-form" loading={createMutation.isPending}>
          Tạo nhân viên
        </Button>
      </div>
    </div>
  );
}
