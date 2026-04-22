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
import { employeeService, type Dependent, type CreateDependentRequest } from '@/services/hrm/employee';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Users, Pencil, Trash2, Plus } from 'lucide-react';

const RELATIONSHIP_LABELS: Record<string, string> = {
  SPOUSE: 'Vợ/Chồng', CHILD: 'Con', PARENT: 'Cha/Mẹ', SIBLING: 'Anh/Chị/Em', OTHER: 'Khác',
};

const depSchema = z.object({
  full_name: z.string().min(1, 'Bắt buộc').max(200),
  relationship: z.enum(['SPOUSE', 'CHILD', 'PARENT', 'SIBLING', 'OTHER'], { required_error: 'Bắt buộc' }),
  date_of_birth: z.string().optional(),
  cccd_or_birth_cert: z.string().optional(),
  tax_deduction_registered: z.boolean().optional(),
  tax_deduction_from: z.string().optional(),
  tax_deduction_to: z.string().optional(),
  notes: z.string().optional(),
});

type DepFormData = z.infer<typeof depSchema>;

interface DepFormProps {
  formId: string;
  initial?: Partial<DepFormData>;
  onSubmit: (data: CreateDependentRequest) => void;
}

function DependentForm({ formId, initial, onSubmit }: DepFormProps) {
  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<DepFormData>({
    resolver: zodResolver(depSchema),
    defaultValues: {
      full_name: initial?.full_name ?? '',
      relationship: initial?.relationship,
      date_of_birth: initial?.date_of_birth ?? '',
      cccd_or_birth_cert: initial?.cccd_or_birth_cert ?? '',
      tax_deduction_registered: initial?.tax_deduction_registered ?? false,
      notes: initial?.notes ?? '',
    },
  });

  return (
    <form id={formId} onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Họ và tên *</Label>
          <Input {...register('full_name')} />
          {errors.full_name && <p className="text-xs text-danger">{errors.full_name.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Quan hệ *</Label>
          <Select value={watch('relationship') ?? ''} onValueChange={(v) => setValue('relationship', v as DepFormData['relationship'])}>
            <SelectTrigger><SelectValue placeholder="Chọn..." /></SelectTrigger>
            <SelectContent>
              {Object.entries(RELATIONSHIP_LABELS).map(([k, v]) => (
                <SelectItem key={k} value={k}>{v}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.relationship && <p className="text-xs text-danger">{errors.relationship.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày sinh</Label>
          <Input {...register('date_of_birth')} type="date" />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Số CCCD / Giấy khai sinh</Label>
          <Input {...register('cccd_or_birth_cert')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Giảm trừ gia cảnh từ</Label>
          <Input {...register('tax_deduction_from')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Giảm trừ gia cảnh đến</Label>
          <Input {...register('tax_deduction_to')} type="date" />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Ghi chú</Label>
          <Input {...register('notes')} />
        </div>
      </div>
    </form>
  );
}

interface DependentSectionProps {
  employeeId: string;
  canWrite: boolean;
}

export function DependentSection({ employeeId, canWrite }: DependentSectionProps) {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<Dependent | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Dependent | null>(null);

  const { data: deps = [], isLoading } = useQuery({
    queryKey: ['hrm', 'dependents', employeeId],
    queryFn: () => employeeService.listDependents(employeeId),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateDependentRequest) => employeeService.createDependent(employeeId, data),
    onSuccess: () => {
      toast('Thêm người phụ thuộc thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'dependents', employeeId] });
      setCreateOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateDependentRequest> }) =>
      employeeService.updateDependent(employeeId, id, data),
    onSuccess: () => {
      toast('Cập nhật người phụ thuộc thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'dependents', employeeId] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => employeeService.deleteDependent(employeeId, id),
    onSuccess: () => {
      toast('Đã xóa người phụ thuộc', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'dependents', employeeId] });
      setDeleteTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-text-primary">Người phụ thuộc</h3>
        {canWrite && (
          <Button size="sm" variant="outline" onClick={() => setCreateOpen(true)}>
            <Plus className="w-4 h-4" />Thêm
          </Button>
        )}
      </div>

      {isLoading ? (
        <p className="text-sm text-text-secondary">Đang tải...</p>
      ) : deps.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
          <Users className="w-8 h-8 opacity-30" />
          <p className="text-sm">Chưa có người phụ thuộc</p>
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Họ và tên</TableHead>
              <TableHead>Quan hệ</TableHead>
              <TableHead>Ngày sinh</TableHead>
              <TableHead>Giảm trừ thuế</TableHead>
              {canWrite && <TableHead className="w-20"></TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {deps.map((dep) => (
              <TableRow key={dep.id}>
                <TableCell className="font-medium">{dep.full_name}</TableCell>
                <TableCell>
                  <Badge variant="outline">{RELATIONSHIP_LABELS[dep.relationship] ?? dep.relationship}</Badge>
                </TableCell>
                <TableCell className="text-sm">{dep.date_of_birth ? formatDate(dep.date_of_birth) : '–'}</TableCell>
                <TableCell>
                  {dep.tax_deduction_registered && (
                    <Badge variant="default" className="text-xs">Đăng ký</Badge>
                  )}
                </TableCell>
                {canWrite && (
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" onClick={() => setEditTarget(dep)}>
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon" className="text-danger hover:text-danger"
                        onClick={() => setDeleteTarget(dep)}>
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      {/* Create dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Thêm người phụ thuộc</DialogTitle></DialogHeader>
          <DependentForm formId="dep-create-form" onSubmit={(data) => createMutation.mutate(data)} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="dep-create-form" loading={createMutation.isPending}>Thêm</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      {editTarget && (
        <Dialog open={!!editTarget} onOpenChange={(v) => !v && setEditTarget(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Chỉnh sửa người phụ thuộc</DialogTitle></DialogHeader>
            <DependentForm
              formId="dep-edit-form"
              initial={editTarget as Partial<DepFormData>}
              onSubmit={(data) => updateMutation.mutate({ id: editTarget.id, data })}
            />
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditTarget(null)}>Hủy</Button>
              <Button type="submit" form="dep-edit-form" loading={updateMutation.isPending}>Lưu</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Delete confirm */}
      {deleteTarget && (
        <Dialog open={!!deleteTarget} onOpenChange={(v) => !v && setDeleteTarget(null)}>
          <DialogContent>
            <DialogHeader><DialogTitle>Xác nhận xóa</DialogTitle></DialogHeader>
            <p className="text-sm text-text-secondary">
              Bạn có chắc muốn xóa người phụ thuộc <strong>{deleteTarget.full_name}</strong>?
            </p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDeleteTarget(null)}>Hủy</Button>
              <Button
                className="bg-danger text-white hover:bg-danger/90"
                loading={deleteMutation.isPending}
                onClick={() => deleteMutation.mutate(deleteTarget.id)}
              >
                Xóa
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
