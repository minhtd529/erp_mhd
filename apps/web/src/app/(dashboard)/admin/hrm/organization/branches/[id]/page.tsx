'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { BranchForm } from '@/components/hrm/branch-form';
import { AssignHeadDialog } from '@/components/hrm/assign-head-dialog';
import { branchService, type UpdateBranchRequest } from '@/services/hrm/organization';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { ArrowLeft, Pencil, UserCog, PowerOff } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO'];

function InfoRow({ label, value }: { label: string; value?: string | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value || '–'}</span>
    </div>
  );
}

export default function BranchDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [editOpen, setEditOpen] = React.useState(false);
  const [assignOpen, setAssignOpen] = React.useState(false);
  const [deactivateConfirm, setDeactivateConfirm] = React.useState(false);

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));
  const isHoB = user?.roles?.includes('HEAD_OF_BRANCH') && !canWrite;
  const isOwnBranch = isHoB && user?.branch_id === id;
  const canEdit = canWrite || isOwnBranch;

  const { data: branch, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'branches', id],
    queryFn: () => branchService.getById(id),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateBranchRequest) => branchService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật chi nhánh thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'branches'] });
      setEditOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const assignMutation = useMutation({
    mutationFn: (userId: string) => branchService.assignHead(id, userId),
    onSuccess: () => {
      toast('Đã phân công trưởng chi nhánh', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'branches', id] });
      setAssignOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deactivateMutation = useMutation({
    mutationFn: () => branchService.deactivate(id),
    onSuccess: () => {
      toast('Đã vô hiệu hóa chi nhánh', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'branches'] });
      router.push('/admin/hrm/organization/branches');
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError || !branch) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không tìm thấy chi nhánh.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 max-w-3xl">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/organization/branches')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-semibold text-text-primary">{branch.name}</h1>
            <Badge variant={branch.is_active ? 'default' : 'secondary'}>
              {branch.is_active ? 'Hoạt động' : 'Ngưng'}
            </Badge>
            {branch.is_head_office && (
              <Badge variant="outline" className="text-secondary border-secondary">Trụ sở chính</Badge>
            )}
          </div>
          <p className="text-xs text-text-secondary font-mono mt-0.5">{branch.code}</p>
        </div>
        <div className="flex gap-2">
          {canEdit && (
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="w-4 h-4" />Chỉnh sửa
            </Button>
          )}
          {canWrite && (
            <Button variant="outline" onClick={() => setAssignOpen(true)}>
              <UserCog className="w-4 h-4" />Phân công trưởng CN
            </Button>
          )}
          {canWrite && branch.is_active && (
            <Button variant="outline" className="text-danger border-danger hover:bg-danger/5" onClick={() => setDeactivateConfirm(true)}>
              <PowerOff className="w-4 h-4" />Vô hiệu hóa
            </Button>
          )}
        </div>
      </div>

      {/* Info card */}
      <Card>
        <CardContent className="p-5 grid grid-cols-2 gap-4">
          <InfoRow label="Địa chỉ" value={branch.address} />
          <InfoRow label="Thành phố" value={branch.city} />
          <InfoRow label="Điện thoại" value={branch.phone} />
          <InfoRow label="Mã số thuế" value={branch.tax_code} />
          <InfoRow label="Ngày thành lập" value={branch.established_date} />
          <InfoRow label="Trưởng chi nhánh (UserID)" value={branch.head_of_branch_user_id} />
          <InfoRow label="Số quyết định ủy quyền" value={branch.authorization_doc_number} />
          <InfoRow label="Ngày ủy quyền" value={branch.authorization_date} />
          <InfoRow label="Ngày tạo" value={formatDate(branch.created_at)} />
          <InfoRow label="Cập nhật lần cuối" value={formatDate(branch.updated_at)} />
        </CardContent>
      </Card>

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Chỉnh sửa chi nhánh</DialogTitle></DialogHeader>
          <BranchForm
            formId="branch-edit-form"
            initial={branch}
            canEditCritical={canWrite}
            onSubmit={(data) => updateMutation.mutate(data)}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Hủy</Button>
            <Button type="submit" form="branch-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Assign head dialog */}
      <AssignHeadDialog
        open={assignOpen}
        onOpenChange={setAssignOpen}
        title="Phân công Trưởng chi nhánh"
        label="User ID (UUID)"
        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
        loading={assignMutation.isPending}
        onSubmit={(userId) => assignMutation.mutate(userId)}
      />

      {/* Deactivate confirm */}
      <Dialog open={deactivateConfirm} onOpenChange={setDeactivateConfirm}>
        <DialogContent>
          <DialogHeader><DialogTitle>Xác nhận vô hiệu hóa</DialogTitle></DialogHeader>
          <p className="text-sm text-text-secondary">
            Bạn có chắc muốn vô hiệu hóa chi nhánh <strong>{branch.name}</strong>?
            Hành động này không thể hoàn tác qua UI.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeactivateConfirm(false)}>Hủy</Button>
            <Button
              className="bg-danger text-white hover:bg-danger/90"
              loading={deactivateMutation.isPending}
              onClick={() => deactivateMutation.mutate()}
            >
              Vô hiệu hóa
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
