'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { employeeService, type CreateSalaryHistoryRequest } from '@/services/hrm/employee';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate, formatCurrency } from '@/lib/utils';
import { TrendingUp, Plus } from 'lucide-react';

const CHANGE_TYPE_LABELS: Record<string, string> = {
  INITIAL: 'Khởi tạo', INCREASE: 'Tăng lương', DECREASE: 'Giảm lương',
  PROMOTION: 'Thăng chức', ADJUSTMENT: 'Điều chỉnh',
};

const CHANGE_TYPE_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  INITIAL: 'outline', INCREASE: 'default', DECREASE: 'secondary',
  PROMOTION: 'default', ADJUSTMENT: 'outline',
};

const salaryHistorySchema = z.object({
  effective_from: z.string().min(1, 'Bắt buộc'),
  base_salary: z.string().min(1, 'Bắt buộc'),
  allowances_total: z.string().optional(),
  change_type: z.enum(['INITIAL', 'INCREASE', 'DECREASE', 'PROMOTION', 'ADJUSTMENT'], { required_error: 'Bắt buộc' }),
  reason: z.string().optional(),
});

type SalaryHistoryFormData = z.infer<typeof salaryHistorySchema>;

interface SalaryHistorySectionProps {
  employeeId: string;
  canRead: boolean;
  canWrite: boolean;
}

export function SalaryHistorySection({ employeeId, canRead, canWrite }: SalaryHistorySectionProps) {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = React.useState(false);

  const { data: history = [], isLoading } = useQuery({
    queryKey: ['hrm', 'salary-history', employeeId],
    queryFn: () => employeeService.listSalaryHistory(employeeId),
    enabled: canRead,
  });

  const { register, handleSubmit, watch, setValue, reset, formState: { errors } } = useForm<SalaryHistoryFormData>({
    resolver: zodResolver(salaryHistorySchema),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateSalaryHistoryRequest) => employeeService.createSalaryHistory(employeeId, data),
    onSuccess: () => {
      toast('Thêm lịch sử lương thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'salary-history', employeeId] });
      setCreateOpen(false);
      reset();
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const submit = (data: SalaryHistoryFormData) => {
    createMutation.mutate({
      ...data,
      allowances_total: data.allowances_total || undefined,
      reason: data.reason || undefined,
    });
  };

  if (!canRead) {
    return (
      <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
        <TrendingUp className="w-8 h-8 opacity-30" />
        <p className="text-sm">Bạn không có quyền xem lịch sử lương</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-text-primary">Lịch sử lương</h3>
        {canWrite && (
          <Button size="sm" variant="outline" onClick={() => setCreateOpen(true)}>
            <Plus className="w-4 h-4" />Thêm
          </Button>
        )}
      </div>

      {isLoading ? (
        <p className="text-sm text-text-secondary">Đang tải...</p>
      ) : history.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
          <TrendingUp className="w-8 h-8 opacity-30" />
          <p className="text-sm">Chưa có lịch sử lương</p>
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hiệu lực từ</TableHead>
              <TableHead>Lương cơ bản</TableHead>
              <TableHead>Phụ cấp</TableHead>
              <TableHead>Loại thay đổi</TableHead>
              <TableHead>Lý do</TableHead>
              <TableHead>Người tạo</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {history.map((h) => (
              <TableRow key={h.id}>
                <TableCell className="text-sm">{formatDate(h.effective_from)}</TableCell>
                <TableCell className="font-mono text-sm">
                  {formatCurrency(parseFloat(h.base_salary))}
                </TableCell>
                <TableCell className="font-mono text-sm text-text-secondary">
                  {h.allowances_total ? formatCurrency(parseFloat(h.allowances_total)) : '–'}
                </TableCell>
                <TableCell>
                  <Badge variant={CHANGE_TYPE_VARIANTS[h.change_type] ?? 'outline'}>
                    {CHANGE_TYPE_LABELS[h.change_type] ?? h.change_type}
                  </Badge>
                </TableCell>
                <TableCell className="text-sm text-text-secondary">{h.reason ?? '–'}</TableCell>
                <TableCell className="text-sm text-text-secondary">{h.created_by_name ?? '–'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={(v) => { setCreateOpen(v); if (!v) reset(); }}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Thêm lịch sử lương</DialogTitle></DialogHeader>
          <form id="salary-create-form" onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <Label>Hiệu lực từ ngày *</Label>
              <Input {...register('effective_from')} type="date" />
              {errors.effective_from && <p className="text-xs text-danger">{errors.effective_from.message}</p>}
            </div>
            <div className="flex flex-col gap-1">
              <Label>Lương cơ bản (VNĐ) *</Label>
              <Input {...register('base_salary')} placeholder="ví dụ: 15000000" />
              {errors.base_salary && <p className="text-xs text-danger">{errors.base_salary.message}</p>}
            </div>
            <div className="flex flex-col gap-1">
              <Label>Tổng phụ cấp (VNĐ)</Label>
              <Input {...register('allowances_total')} placeholder="ví dụ: 3000000" />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Loại thay đổi *</Label>
              <Select value={watch('change_type') ?? ''} onValueChange={(v) => setValue('change_type', v as SalaryHistoryFormData['change_type'])}>
                <SelectTrigger><SelectValue placeholder="Chọn..." /></SelectTrigger>
                <SelectContent>
                  {Object.entries(CHANGE_TYPE_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {errors.change_type && <p className="text-xs text-danger">{errors.change_type.message}</p>}
            </div>
            <div className="flex flex-col gap-1">
              <Label>Lý do</Label>
              <Input {...register('reason')} />
            </div>
          </form>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="salary-create-form" loading={createMutation.isPending}>Thêm</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
