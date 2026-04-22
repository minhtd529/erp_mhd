'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { employeeService, type SensitiveFields, type UpdateSensitiveRequest } from '@/services/hrm/employee';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDateTime } from '@/lib/utils';
import { ShieldAlert, Eye, Pencil } from 'lucide-react';

interface SensitiveModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  employeeId: string;
  employeeName: string;
}

function InfoRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm font-mono text-text-primary">{value || '–'}</span>
    </div>
  );
}

const sensitiveSchema = z.object({
  cccd: z.string().optional(),
  cccd_issued_date: z.string().optional(),
  cccd_issued_place: z.string().optional(),
  passport_number: z.string().optional(),
  passport_expiry: z.string().optional(),
  mst_ca_nhan: z.string().optional(),
  so_bhxh: z.string().optional(),
  bank_account: z.string().optional(),
  bank_name: z.string().optional(),
  bank_branch: z.string().optional(),
});

type SensitiveFormData = z.infer<typeof sensitiveSchema>;

// ─── Confirmation gate ────────────────────────────────────────────────────────

function ConfirmAccessDialog({ open, onConfirm, onCancel }: {
  open: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onCancel()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-danger">
            <ShieldAlert className="w-5 h-5" />
            Truy cập dữ liệu nhạy cảm
          </DialogTitle>
        </DialogHeader>
        <p className="text-sm text-text-secondary leading-relaxed">
          Thao tác này sẽ giải mã và hiển thị thông tin cá nhân nhạy cảm (CCCD, MST, BHXH, tài khoản ngân hàng).
          <strong className="text-text-primary"> Mọi truy cập đều được ghi lại trong nhật ký kiểm toán.</strong>
        </p>
        <p className="text-sm text-text-secondary">Bạn có chắc muốn tiếp tục?</p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>Hủy</Button>
          <Button className="bg-danger text-white hover:bg-danger/90" onClick={onConfirm}>
            <Eye className="w-4 h-4" />Tôi hiểu — xem dữ liệu
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ─── Edit dialog ──────────────────────────────────────────────────────────────

function SensitiveEditDialog({ open, onOpenChange, employeeId, initial }: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  employeeId: string;
  initial?: SensitiveFields;
}) {
  const qc = useQueryClient();
  const { register, handleSubmit, reset } = useForm<SensitiveFormData>({
    resolver: zodResolver(sensitiveSchema),
    defaultValues: {
      cccd: initial?.cccd ?? '',
      cccd_issued_date: initial?.cccd_issued_date ?? '',
      cccd_issued_place: initial?.cccd_issued_place ?? '',
      passport_number: initial?.passport_number ?? '',
      passport_expiry: initial?.passport_expiry ?? '',
      mst_ca_nhan: initial?.mst_ca_nhan ?? '',
      so_bhxh: initial?.so_bhxh ?? '',
      bank_account: initial?.bank_account ?? '',
      bank_name: initial?.bank_name ?? '',
      bank_branch: initial?.bank_branch ?? '',
    },
  });

  React.useEffect(() => {
    if (open && initial) {
      reset({
        cccd: initial.cccd ?? '',
        cccd_issued_date: initial.cccd_issued_date ?? '',
        cccd_issued_place: initial.cccd_issued_place ?? '',
        passport_number: initial.passport_number ?? '',
        passport_expiry: initial.passport_expiry ?? '',
        mst_ca_nhan: initial.mst_ca_nhan ?? '',
        so_bhxh: initial.so_bhxh ?? '',
        bank_account: initial.bank_account ?? '',
        bank_name: initial.bank_name ?? '',
        bank_branch: initial.bank_branch ?? '',
      });
    }
  }, [open, initial, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: UpdateSensitiveRequest) => employeeService.updateSensitive(employeeId, data),
    onSuccess: () => {
      toast('Cập nhật thông tin nhạy cảm thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'sensitive', employeeId] });
      onOpenChange(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const submit = (data: SensitiveFormData) => {
    const req: UpdateSensitiveRequest = {};
    if (data.cccd) req.cccd = data.cccd;
    if (data.cccd_issued_date) req.cccd_issued_date = data.cccd_issued_date;
    if (data.cccd_issued_place) req.cccd_issued_place = data.cccd_issued_place;
    if (data.passport_number) req.passport_number = data.passport_number;
    if (data.passport_expiry) req.passport_expiry = data.passport_expiry;
    if (data.mst_ca_nhan) req.mst_ca_nhan = data.mst_ca_nhan;
    if (data.so_bhxh) req.so_bhxh = data.so_bhxh;
    if (data.bank_account) req.bank_account = data.bank_account;
    if (data.bank_name) req.bank_name = data.bank_name;
    if (data.bank_branch) req.bank_branch = data.bank_branch;
    updateMutation.mutate(req);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Cập nhật thông tin nhạy cảm</DialogTitle>
        </DialogHeader>
        <form id="sensitive-edit-form" onSubmit={handleSubmit(submit)} className="flex flex-col gap-3">
          <p className="text-xs text-text-secondary bg-warning/10 border border-warning/30 rounded p-2">
            Chỉ điền các trường cần cập nhật. Trường để trống sẽ không bị thay đổi.
          </p>
          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1 col-span-2">
              <Label>Số CCCD/CMND</Label>
              <Input {...register('cccd')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ngày cấp CCCD</Label>
              <Input {...register('cccd_issued_date')} type="date" />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Nơi cấp CCCD</Label>
              <Input {...register('cccd_issued_place')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số hộ chiếu</Label>
              <Input {...register('passport_number')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Ngày hết hạn hộ chiếu</Label>
              <Input {...register('passport_expiry')} type="date" />
            </div>
            <div className="flex flex-col gap-1">
              <Label>MST cá nhân</Label>
              <Input {...register('mst_ca_nhan')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số BHXH</Label>
              <Input {...register('so_bhxh')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Số tài khoản ngân hàng</Label>
              <Input {...register('bank_account')} />
            </div>
            <div className="flex flex-col gap-1">
              <Label>Tên ngân hàng</Label>
              <Input {...register('bank_name')} />
            </div>
            <div className="flex flex-col gap-1 col-span-2">
              <Label>Chi nhánh ngân hàng</Label>
              <Input {...register('bank_branch')} />
            </div>
          </div>
        </form>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Hủy</Button>
          <Button type="submit" form="sensitive-edit-form" loading={updateMutation.isPending}>Lưu</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// ─── Main modal ───────────────────────────────────────────────────────────────

export function SensitiveModal({ open, onOpenChange, employeeId, employeeName }: SensitiveModalProps) {
  const [confirmed, setConfirmed] = React.useState(false);
  const [editOpen, setEditOpen] = React.useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['hrm', 'sensitive', employeeId],
    queryFn: () => employeeService.getSensitive(employeeId),
    enabled: open && confirmed,
    staleTime: 0,
  });

  React.useEffect(() => {
    if (!open) {
      setConfirmed(false);
    }
  }, [open]);

  if (!confirmed) {
    return (
      <ConfirmAccessDialog
        open={open}
        onConfirm={() => setConfirmed(true)}
        onCancel={() => onOpenChange(false)}
      />
    );
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <ShieldAlert className="w-5 h-5 text-danger" />
              Thông tin nhạy cảm — {employeeName}
            </DialogTitle>
          </DialogHeader>

          {isLoading && <PageSpinner />}

          {data && (
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Số CCCD/CMND" value={data.cccd} />
              <InfoRow label="Ngày cấp CCCD" value={data.cccd_issued_date} />
              <InfoRow label="Nơi cấp CCCD" value={data.cccd_issued_place} />
              <InfoRow label="Số hộ chiếu" value={data.passport_number} />
              <InfoRow label="Hết hạn hộ chiếu" value={data.passport_expiry} />
              <InfoRow label="MST cá nhân" value={data.mst_ca_nhan} />
              <InfoRow label="Số BHXH" value={data.so_bhxh} />
              <InfoRow label="Số tài khoản NH" value={data.bank_account} />
              <InfoRow label="Tên ngân hàng" value={data.bank_name} />
              <InfoRow label="Chi nhánh NH" value={data.bank_branch} />
              <div className="col-span-2 text-xs text-text-secondary border-t pt-2">
                Truy cập lúc: {formatDateTime(data.accessed_at)}
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="w-4 h-4" />Cập nhật
            </Button>
            <Button variant="outline" onClick={() => onOpenChange(false)}>Đóng</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <SensitiveEditDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        employeeId={employeeId}
        initial={data}
      />
    </>
  );
}
