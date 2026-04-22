'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { provisioningService } from '@/services/hrm/provisioning';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { MODULE_ROLES } from '@/lib/roles';
import { ArrowLeft, CheckCircle, XCircle, Play, Ban, AlertTriangle } from 'lucide-react';

const STATUS_LABELS: Record<string, string> = {
  PENDING:   'Chờ duyệt',
  APPROVED:  'Đã duyệt',
  REJECTED:  'Từ chối',
  EXECUTED:  'Đã thực thi',
  CANCELLED: 'Đã hủy',
};

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  PENDING:   'outline',
  APPROVED:  'default',
  REJECTED:  'secondary',
  EXECUTED:  'default',
  CANCELLED: 'secondary',
};

const ROLE_LABELS: Record<string, string> = {
  SUPER_ADMIN:    'Quản trị viên',
  CHAIRMAN:       'Chủ tịch',
  CEO:            'Tổng GĐ',
  HR_MANAGER:     'Trưởng NS',
  HR_STAFF:       'NV Nhân sự',
  HEAD_OF_BRANCH: 'Trưởng CN',
  FIRM_PARTNER:   'Partner',
  AUDIT_MANAGER:  'QL Kiểm toán',
  SENIOR_AUDITOR: 'KTV Cao cấp',
  JUNIOR_AUDITOR: 'Kiểm toán viên',
  AUDIT_STAFF:    'NV Kiểm toán',
  ACCOUNTANT:     'Kế toán',
};

function InfoRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value ?? '–'}</span>
    </div>
  );
}

export default function ProvisioningDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [rejectOpen, setRejectOpen] = React.useState(false);
  const [rejectType, setRejectType] = React.useState<'branch' | 'hr'>('branch');
  const [rejectReason, setRejectReason] = React.useState('');
  const [executeOpen, setExecuteOpen] = React.useState(false);
  const [execFullName, setExecFullName] = React.useState('');
  const [execEmail, setExecEmail] = React.useState('');
  const [tempPassword, setTempPassword] = React.useState('');

  const isBranch   = MODULE_ROLES.hrmProvisioningBranch.some(r => user?.roles?.includes(r));
  const isHR       = MODULE_ROLES.hrmProvisioningHR.some(r => user?.roles?.includes(r));
  const canExecute = MODULE_ROLES.hrmProvisioningExecute.some(r => user?.roles?.includes(r));
  const canCancel  = MODULE_ROLES.hrmProvisioningCreate.some(r => user?.roles?.includes(r));

  const { data: req, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'provisioning', id],
    queryFn: () => provisioningService.getById(id),
    enabled: !!id,
  });

  function invalidate() {
    qc.invalidateQueries({ queryKey: ['hrm', 'provisioning', id] });
    qc.invalidateQueries({ queryKey: ['hrm', 'provisioning'] });
  }

  const branchApproveMutation = useMutation({
    mutationFn: () => provisioningService.branchApprove(id),
    onSuccess: () => { toast('Đã duyệt (chi nhánh)', 'success'); invalidate(); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const hrApproveMutation = useMutation({
    mutationFn: () => provisioningService.hrApprove(id),
    onSuccess: () => { toast('Đã duyệt (nhân sự)', 'success'); invalidate(); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const rejectMutation = useMutation({
    mutationFn: (reason: string) =>
      rejectType === 'branch'
        ? provisioningService.branchReject(id, { reason })
        : provisioningService.hrReject(id, { reason }),
    onSuccess: () => {
      toast('Đã từ chối yêu cầu', 'success');
      setRejectOpen(false);
      setRejectReason('');
      invalidate();
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const executeMutation = useMutation({
    mutationFn: () => provisioningService.execute(id, { full_name: execFullName, email: execEmail }),
    onSuccess: (res) => {
      setTempPassword(res.temp_password);
      setExecuteOpen(false);
      invalidate();
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const cancelMutation = useMutation({
    mutationFn: () => provisioningService.cancel(id),
    onSuccess: () => { toast('Đã hủy yêu cầu', 'success'); invalidate(); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError || !req) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải yêu cầu.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const isPending = req.status === 'PENDING';
  const isApproved = req.status === 'APPROVED';
  const needsBranchApproval = !!req.requested_branch_id && !req.branch_approved_at;

  return (
    <div className="flex flex-col gap-4 max-w-2xl">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/provisioning')}>
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-text-primary">Chi tiết yêu cầu cấp tài khoản</h1>
            <p className="text-sm text-text-secondary font-mono">{req.id}</p>
          </div>
        </div>
        <Badge variant={STATUS_VARIANTS[req.status] ?? 'outline'}>
          {STATUS_LABELS[req.status] ?? req.status}
        </Badge>
      </div>

      {tempPassword && (
        <div className="rounded-md border border-success/40 bg-success/5 p-4 flex flex-col gap-1">
          <p className="text-sm font-semibold text-success">Tài khoản đã được tạo thành công!</p>
          <p className="text-xs text-text-secondary">Mật khẩu tạm thời (hiển thị một lần):</p>
          <code className="text-sm font-mono bg-surface-raised px-2 py-1 rounded select-all">{tempPassword}</code>
          <p className="text-xs text-text-secondary">Người dùng nên đổi mật khẩu ngay sau khi đăng nhập lần đầu.</p>
        </div>
      )}

      <Card>
        <CardContent className="p-5">
          <div className="grid grid-cols-2 gap-4">
            <InfoRow label="Nhân viên ID" value={req.employee_id} />
            <InfoRow label="Vai trò yêu cầu" value={ROLE_LABELS[req.requested_role] ?? req.requested_role} />
            <InfoRow label="Người tạo" value={req.requested_by} />
            <InfoRow label="Hết hạn" value={formatDate(req.expires_at)} />
            <InfoRow label="Ngày tạo" value={formatDate(req.created_at)} />
            <InfoRow label="Cập nhật" value={formatDate(req.updated_at)} />
            {req.notes && <div className="col-span-2"><InfoRow label="Ghi chú" value={req.notes} /></div>}
          </div>
        </CardContent>
      </Card>

      {req.is_emergency && (
        <Card>
          <CardContent className="p-4 flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-warning flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-text-primary">Yêu cầu khẩn cấp</p>
              {req.emergency_reason && (
                <p className="text-sm text-text-secondary mt-0.5">{req.emergency_reason}</p>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {(req.requested_branch_id || req.branch_approver_id) && (
        <Card>
          <CardContent className="p-5">
            <p className="text-xs font-semibold uppercase tracking-wider text-text-secondary mb-3">Duyệt Chi nhánh</p>
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Chi nhánh" value={req.requested_branch_id ?? '–'} />
              <InfoRow label="Người duyệt" value={req.branch_approver_id ?? '–'} />
              <InfoRow label="Ngày duyệt" value={req.branch_approved_at ? formatDate(req.branch_approved_at) : undefined} />
              {req.branch_rejection_reason && (
                <div className="col-span-2">
                  <InfoRow label="Lý do từ chối" value={req.branch_rejection_reason} />
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {(req.hr_approver_id || req.hr_approved_at) && (
        <Card>
          <CardContent className="p-5">
            <p className="text-xs font-semibold uppercase tracking-wider text-text-secondary mb-3">Duyệt Nhân sự</p>
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Người duyệt NS" value={req.hr_approver_id ?? '–'} />
              <InfoRow label="Ngày duyệt NS" value={req.hr_approved_at ? formatDate(req.hr_approved_at) : undefined} />
              {req.hr_rejection_reason && (
                <div className="col-span-2">
                  <InfoRow label="Lý do từ chối" value={req.hr_rejection_reason} />
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {req.executed_at && (
        <Card>
          <CardContent className="p-5">
            <p className="text-xs font-semibold uppercase tracking-wider text-text-secondary mb-3">Thực thi</p>
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Người thực thi" value={req.executed_by ?? '–'} />
              <InfoRow label="Ngày thực thi" value={formatDate(req.executed_at)} />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Action buttons */}
      {(isPending || isApproved) && (
        <div className="flex flex-wrap gap-2 justify-end">
          {isPending && isBranch && needsBranchApproval && (
            <>
              <Button
                variant="outline"
                className="text-danger border-danger/30 hover:bg-danger/5"
                onClick={() => { setRejectType('branch'); setRejectOpen(true); }}
              >
                <XCircle className="w-4 h-4" />Từ chối (CN)
              </Button>
              <Button onClick={() => branchApproveMutation.mutate()} loading={branchApproveMutation.isPending}>
                <CheckCircle className="w-4 h-4" />Duyệt (CN)
              </Button>
            </>
          )}

          {isPending && isHR && (!req.requested_branch_id || req.branch_approved_at) && (
            <>
              <Button
                variant="outline"
                className="text-danger border-danger/30 hover:bg-danger/5"
                onClick={() => { setRejectType('hr'); setRejectOpen(true); }}
              >
                <XCircle className="w-4 h-4" />Từ chối (NS)
              </Button>
              <Button onClick={() => hrApproveMutation.mutate()} loading={hrApproveMutation.isPending}>
                <CheckCircle className="w-4 h-4" />Duyệt (NS)
              </Button>
            </>
          )}

          {isApproved && canExecute && (
            <Button onClick={() => setExecuteOpen(true)}>
              <Play className="w-4 h-4" />Thực thi — tạo tài khoản
            </Button>
          )}

          {isPending && canCancel && (
            <Button
              variant="outline"
              className="text-text-secondary"
              onClick={() => cancelMutation.mutate()}
              loading={cancelMutation.isPending}
            >
              <Ban className="w-4 h-4" />Hủy yêu cầu
            </Button>
          )}
        </div>
      )}

      {/* Reject dialog */}
      <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Từ chối yêu cầu</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-2">
            <label className="text-sm font-medium text-text-primary">
              Lý do từ chối <span className="text-danger">*</span>
            </label>
            <textarea
              value={rejectReason}
              onChange={e => setRejectReason(e.target.value)}
              rows={3}
              placeholder="Mô tả lý do từ chối..."
              className="rounded-md border border-border bg-surface-paper px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-action/30 resize-none"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejectOpen(false)}>Hủy</Button>
            <Button
              className="bg-danger hover:bg-danger/90 text-white"
              disabled={!rejectReason.trim()}
              loading={rejectMutation.isPending}
              onClick={() => rejectMutation.mutate(rejectReason)}
            >
              Xác nhận từ chối
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Execute dialog */}
      <Dialog open={executeOpen} onOpenChange={setExecuteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Tạo tài khoản người dùng</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4 py-2">
            <p className="text-sm text-text-secondary">
              Điền thông tin để tạo tài khoản hệ thống cho nhân viên.
            </p>
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-text-primary">
                Họ và tên <span className="text-danger">*</span>
              </label>
              <input
                type="text"
                value={execFullName}
                onChange={e => setExecFullName(e.target.value)}
                placeholder="Nguyễn Văn A"
                className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm focus:outline-none focus:ring-2 focus:ring-action/30"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-text-primary">
                Email <span className="text-danger">*</span>
              </label>
              <input
                type="email"
                value={execEmail}
                onChange={e => setExecEmail(e.target.value)}
                placeholder="user@company.com"
                className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm focus:outline-none focus:ring-2 focus:ring-action/30"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setExecuteOpen(false)}>Hủy</Button>
            <Button
              disabled={!execFullName.trim() || !execEmail.trim()}
              loading={executeMutation.isPending}
              onClick={() => executeMutation.mutate()}
            >
              Tạo tài khoản
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
