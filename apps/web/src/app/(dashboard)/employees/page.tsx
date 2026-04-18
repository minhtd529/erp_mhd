'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { SearchInput } from '@/components/shared/search-input';
import { employeeService, type EmployeeCreateRequest } from '@/services/employees';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Plus, Pencil, Trash2 } from 'lucide-react';

const schema = z.object({
  full_name: z.string().min(1, 'Bắt buộc'),
  email: z.string().email('Email không hợp lệ'),
  phone: z.string().optional(),
  department: z.string().optional(),
  position: z.string().optional(),
  hire_date: z.string().optional(),
  is_salesperson: z.boolean().optional(),
  sales_commission_eligible: z.boolean().optional(),
});

type FormData = z.infer<typeof schema>;

function EmployeeForm({ initial, onSubmit, loading }: { initial?: Partial<FormData>; onSubmit: (d: FormData) => void; loading?: boolean }) {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { is_salesperson: false, sales_commission_eligible: false, ...initial },
  });
  return (
    <form id="employee-form" onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Họ và tên *</Label>
          <Input {...register('full_name')} />
          {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Email *</Label>
          <Input {...register('email')} type="email" />
          {errors.email && <p className="text-xs text-danger">{errors.email.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Điện thoại</Label>
          <Input {...register('phone')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Phòng ban</Label>
          <Input {...register('department')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Chức vụ</Label>
          <Input {...register('position')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày vào làm</Label>
          <Input {...register('hire_date')} type="date" />
        </div>
        <div className="flex flex-col gap-2">
          <label className="flex items-center gap-2 cursor-pointer text-sm">
            <input type="checkbox" {...register('is_salesperson')} className="rounded" />
            Là nhân viên kinh doanh
          </label>
          <label className="flex items-center gap-2 cursor-pointer text-sm">
            <input type="checkbox" {...register('sales_commission_eligible')} className="rounded" />
            Được hưởng hoa hồng
          </label>
        </div>
      </div>
    </form>
  );
}

export default function EmployeesPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');
  const [dialog, setDialog] = React.useState<'create' | 'edit' | null>(null);
  const [selected, setSelected] = React.useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['employees', page, q],
    queryFn: () => employeeService.list({ page, size: 20, q: q || undefined }),
  });

  const selectedEmployee = data?.data.find(e => e.id === selected);

  const createMutation = useMutation({
    mutationFn: (d: EmployeeCreateRequest) => employeeService.create(d),
    onSuccess: () => { toast('Tạo nhân viên thành công', 'success'); qc.invalidateQueries({ queryKey: ['employees'] }); setDialog(null); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: (d: EmployeeCreateRequest) => employeeService.update(selected!, d),
    onSuccess: () => { toast('Cập nhật thành công', 'success'); qc.invalidateQueries({ queryKey: ['employees'] }); setDialog(null); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => employeeService.delete(id),
    onSuccess: () => { toast('Đã xóa nhân viên', 'success'); qc.invalidateQueries({ queryKey: ['employees'] }); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <SearchInput placeholder="Tìm theo tên, email..." className="w-80" onSearch={setQ} />
        <Button onClick={() => setDialog('create')}><Plus className="w-4 h-4" />Thêm nhân viên</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Họ tên</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Phòng ban / Chức vụ</TableHead>
                  <TableHead>Loại</TableHead>
                  <TableHead>Ngày vào làm</TableHead>
                  <TableHead className="w-20"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((emp) => (
                  <TableRow key={emp.id}>
                    <TableCell className="font-medium">{emp.full_name}</TableCell>
                    <TableCell className="text-xs">{emp.email}</TableCell>
                    <TableCell className="text-xs">
                      {[emp.department, emp.position].filter(Boolean).join(' / ') || '-'}
                    </TableCell>
                    <TableCell>
                      {emp.is_salesperson && <Badge variant="success">Kinh doanh</Badge>}
                      {emp.sales_commission_eligible && <Badge variant="default" className="ml-1">Hoa hồng</Badge>}
                    </TableCell>
                    <TableCell className="text-xs">{emp.hire_date ? formatDate(emp.hire_date) : '-'}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" onClick={() => { setSelected(emp.id); setDialog('edit'); }}>
                          <Pencil className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          variant="ghost" size="icon" className="text-danger hover:text-danger"
                          onClick={() => { if (confirm('Xóa nhân viên này?')) deleteMutation.mutate(emp.id); }}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={6} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
      {data && <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />}

      <Dialog open={dialog === 'create'} onOpenChange={(o) => !o && setDialog(null)}>
        <DialogContent><DialogHeader><DialogTitle>Thêm nhân viên mới</DialogTitle></DialogHeader>
          <EmployeeForm onSubmit={(d) => createMutation.mutate(d)} loading={createMutation.isPending} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="employee-form" loading={createMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={dialog === 'edit'} onOpenChange={(o) => !o && setDialog(null)}>
        <DialogContent><DialogHeader><DialogTitle>Chỉnh sửa nhân viên</DialogTitle></DialogHeader>
          {selectedEmployee && (
            <EmployeeForm initial={selectedEmployee} onSubmit={(d) => updateMutation.mutate(d)} loading={updateMutation.isPending} />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialog(null)}>Hủy</Button>
            <Button type="submit" form="employee-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
