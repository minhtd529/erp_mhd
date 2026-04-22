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
import { employeeService, type Contract, type CreateContractRequest } from '@/services/hrm/employee';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate, formatCurrency } from '@/lib/utils';
import { FileText, Pencil, StopCircle, Plus } from 'lucide-react';

const CONTRACT_TYPE_LABELS: Record<string, string> = {
  PROBATION: 'Thử việc', DEFINITE_TERM: 'Có thời hạn',
  INDEFINITE: 'Không thời hạn', INTERN: 'Thực tập',
};

const contractSchema = z.object({
  contract_number: z.string().optional(),
  contract_type: z.enum(['PROBATION', 'DEFINITE_TERM', 'INDEFINITE', 'INTERN'], { required_error: 'Bắt buộc' }),
  start_date: z.string().min(1, 'Bắt buộc'),
  end_date: z.string().optional(),
  signed_date: z.string().optional(),
  salary_at_signing: z.coerce.number().optional(),
  position_at_signing: z.string().optional(),
  notes: z.string().optional(),
});

type ContractFormData = z.infer<typeof contractSchema>;

interface ContractFormProps {
  formId: string;
  initial?: Partial<ContractFormData>;
  onSubmit: (data: CreateContractRequest) => void;
}

function ContractForm({ formId, initial, onSubmit }: ContractFormProps) {
  const { register, handleSubmit, watch, setValue, formState: { errors } } = useForm<ContractFormData>({
    resolver: zodResolver(contractSchema),
    defaultValues: {
      contract_number: initial?.contract_number ?? '',
      contract_type: initial?.contract_type,
      start_date: initial?.start_date ?? '',
      end_date: initial?.end_date ?? '',
      signed_date: initial?.signed_date ?? '',
      salary_at_signing: initial?.salary_at_signing,
      position_at_signing: initial?.position_at_signing ?? '',
      notes: initial?.notes ?? '',
    },
  });

  const submit = (data: ContractFormData) => {
    onSubmit({
      ...data,
      end_date: data.end_date || undefined,
      signed_date: data.signed_date || undefined,
    });
  };

  return (
    <form id={formId} onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <Label>Số hợp đồng</Label>
          <Input {...register('contract_number')} />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Loại hợp đồng *</Label>
          <Select value={watch('contract_type') ?? ''} onValueChange={(v) => setValue('contract_type', v as ContractFormData['contract_type'])}>
            <SelectTrigger><SelectValue placeholder="Chọn..." /></SelectTrigger>
            <SelectContent>
              {Object.entries(CONTRACT_TYPE_LABELS).map(([k, v]) => (
                <SelectItem key={k} value={k}>{v}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.contract_type && <p className="text-xs text-danger">{errors.contract_type.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày bắt đầu *</Label>
          <Input {...register('start_date')} type="date" />
          {errors.start_date && <p className="text-xs text-danger">{errors.start_date.message}</p>}
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày kết thúc</Label>
          <Input {...register('end_date')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Ngày ký</Label>
          <Input {...register('signed_date')} type="date" />
        </div>
        <div className="flex flex-col gap-1">
          <Label>Lương tại thời điểm ký (VNĐ)</Label>
          <Input {...register('salary_at_signing')} type="number" />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Chức vụ tại thời điểm ký</Label>
          <Input {...register('position_at_signing')} />
        </div>
        <div className="flex flex-col gap-1 col-span-2">
          <Label>Ghi chú</Label>
          <Input {...register('notes')} />
        </div>
      </div>
    </form>
  );
}

interface ContractSectionProps {
  employeeId: string;
  canWrite: boolean;
}

export function ContractSection({ employeeId, canWrite }: ContractSectionProps) {
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = React.useState(false);
  const [editTarget, setEditTarget] = React.useState<Contract | null>(null);
  const [terminateTarget, setTerminateTarget] = React.useState<Contract | null>(null);

  const { data: contracts = [], isLoading } = useQuery({
    queryKey: ['hrm', 'contracts', employeeId],
    queryFn: () => employeeService.listContracts(employeeId),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateContractRequest) => employeeService.createContract(employeeId, data),
    onSuccess: () => {
      toast('Tạo hợp đồng thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'contracts', employeeId] });
      setCreateOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateContractRequest> }) =>
      employeeService.updateContract(employeeId, id, data),
    onSuccess: () => {
      toast('Cập nhật hợp đồng thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'contracts', employeeId] });
      setEditTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const terminateMutation = useMutation({
    mutationFn: (id: string) => employeeService.terminateContract(employeeId, id),
    onSuccess: () => {
      toast('Đã chấm dứt hợp đồng', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'contracts', employeeId] });
      setTerminateTarget(null);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-text-primary">Hợp đồng lao động</h3>
        {canWrite && (
          <Button size="sm" variant="outline" onClick={() => setCreateOpen(true)}>
            <Plus className="w-4 h-4" />Thêm hợp đồng
          </Button>
        )}
      </div>

      {isLoading ? (
        <p className="text-sm text-text-secondary">Đang tải...</p>
      ) : contracts.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-10 text-text-secondary">
          <FileText className="w-8 h-8 opacity-30" />
          <p className="text-sm">Chưa có hợp đồng</p>
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Số HĐ</TableHead>
              <TableHead>Loại</TableHead>
              <TableHead>Từ ngày</TableHead>
              <TableHead>Đến ngày</TableHead>
              <TableHead>Trạng thái</TableHead>
              {canWrite && <TableHead className="w-24"></TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {contracts.map((c) => (
              <TableRow key={c.id}>
                <TableCell className="font-mono text-xs">{c.contract_number ?? '–'}</TableCell>
                <TableCell>
                  <Badge variant="outline">{CONTRACT_TYPE_LABELS[c.contract_type] ?? c.contract_type}</Badge>
                </TableCell>
                <TableCell className="text-sm">{formatDate(c.start_date)}</TableCell>
                <TableCell className="text-sm">{c.end_date ? formatDate(c.end_date) : '–'}</TableCell>
                <TableCell>
                  <Badge variant={c.is_current ? 'default' : 'secondary'}>
                    {c.is_current ? 'Hiện tại' : 'Kết thúc'}
                  </Badge>
                </TableCell>
                {canWrite && (
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" onClick={() => setEditTarget(c)}>
                        <Pencil className="w-3.5 h-3.5" />
                      </Button>
                      {c.is_current && (
                        <Button variant="ghost" size="icon" className="text-danger hover:text-danger"
                          onClick={() => setTerminateTarget(c)}>
                          <StopCircle className="w-3.5 h-3.5" />
                        </Button>
                      )}
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
          <DialogHeader><DialogTitle>Thêm hợp đồng</DialogTitle></DialogHeader>
          <ContractForm formId="contract-create-form" onSubmit={(data) => createMutation.mutate(data)} />
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Hủy</Button>
            <Button type="submit" form="contract-create-form" loading={createMutation.isPending}>Tạo</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit dialog */}
      {editTarget && (
        <Dialog open={!!editTarget} onOpenChange={(v) => !v && setEditTarget(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader><DialogTitle>Chỉnh sửa hợp đồng</DialogTitle></DialogHeader>
            <ContractForm
              formId="contract-edit-form"
              initial={editTarget as Partial<ContractFormData>}
              onSubmit={(data) => updateMutation.mutate({ id: editTarget.id, data })}
            />
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditTarget(null)}>Hủy</Button>
              <Button type="submit" form="contract-edit-form" loading={updateMutation.isPending}>Lưu</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Terminate confirm */}
      {terminateTarget && (
        <Dialog open={!!terminateTarget} onOpenChange={(v) => !v && setTerminateTarget(null)}>
          <DialogContent>
            <DialogHeader><DialogTitle>Chấm dứt hợp đồng</DialogTitle></DialogHeader>
            <p className="text-sm text-text-secondary">
              Bạn có chắc muốn chấm dứt hợp đồng <strong>{terminateTarget.contract_number ?? terminateTarget.id}</strong>?
              Hành động này không thể hoàn tác.
            </p>
            <DialogFooter>
              <Button variant="outline" onClick={() => setTerminateTarget(null)}>Hủy</Button>
              <Button
                className="bg-danger text-white hover:bg-danger/90"
                loading={terminateMutation.isPending}
                onClick={() => terminateMutation.mutate(terminateTarget.id)}
              >
                Chấm dứt
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
