'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { DepartmentForm } from '@/components/hrm/department-form';
import { AssignHeadDialog } from '@/components/hrm/assign-head-dialog';
import { departmentService, type UpdateDepartmentRequest } from '@/services/hrm/organization';
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

export default function DepartmentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [editOpen, setEditOpen] = React.useState(false);
  const [assignOpen, setAssignOpen] = React.useState(false);
  const [deactivateConfirm, setDeactivateConfirm] = React.useState(false);

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));

  const { data: dept, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'departments', id],
    queryFn: () => departmentService.getById(id),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateDepartmentRequest) => departmentService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật phòng ban thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'departments'] });
      setEditOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const assignMutation = useMutation({
    mutationFn: (employeeId: string) => departmentService.assignHead(id, employeeId),
    onSuccess: () => {
      toast('Đã phân công trưởng phòng', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'departments', id] });
      setAssignOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const deactivateMutation = useMutation({
    mutationFn: () => departmentService.deactivate(id),
    onSuccess: () => {
      toast('Đã vô hiệu hóa phòng ban', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'departments'] });
      router.push('/admin/hrm/organization/departments');
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError || !dept) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không tìm thấy phòng ban.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 max-w-2xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/organization/departments')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-semibold text-text-primary">{dept.name}</h1>
            <Badge variant={dept.is_active ? 'default' : 'secondary'}>
              {dept.is_active ? 'Hoạt động' : 'Ngưng'}
            </Badge>
            <Badge variant="outline">{dept.dept_type}</Badge>
          </div>
          <p className="text-xs text-text-secondary font-mono mt-0.5">{dept.code}</p>
        </div>
        {canWrite && (
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="w-4 h-4" />Chỉnh sửa
            </Button>
            <Button variant="outline" onClick={() => setAssignOpen(true)}>
              <UserCog className="w-4 h-4" />Phân công trưởng phòng
            </Button>
            {dept.is_active && (
              <Button variant="outline" className="text-danger border-danger hover:bg-danger/5" onClick={() => setDeactivateConfirm(true)}>
                <PowerOff className="w-4 h-4" />Vô hiệu hóa
              </Button>
            )}
          </div>
        )}
      </div>

      <Card>
        <CardContent className="p-5 grid grid-cols-2 gap-4">
          <InfoRow label="Mô tả" value={dept.description} />
          <InfoRow label="Loại phòng ban" value={dept.dept_type} />
          <InfoRow label="Trưởng phòng (Employee ID)" value={dept.head_employee_id} />
          <InfoRow label="Ngày tạo" value={formatDate(dept.created_at)} />
          <InfoRow label="Cập nhật lần cuối" value={formatDate(dept.updated_at)} />
        </CardContent>
      </Card>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>Chỉnh sửa phòng ban</DialogTitle></DialogHeader>
          <DepartmentForm
            formId="dept-edit-form"
            initial={dept}
            onSubmit={(data) => updateMutation.mutate(data)}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Hủy</Button>
            <Button type="submit" form="dept-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AssignHeadDialog
        open={assignOpen}
        onOpenChange={setAssignOpen}
        title="Phân công Trưởng phòng"
        label="Employee ID (UUID)"
        placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
        loading={assignMutation.isPending}
        onSubmit={(empId) => assignMutation.mutate(empId)}
      />

      <Dialog open={deactivateConfirm} onOpenChange={setDeactivateConfirm}>
        <DialogContent>
          <DialogHeader><DialogTitle>Xác nhận vô hiệu hóa</DialogTitle></DialogHeader>
          <p className="text-sm text-text-secondary">
            Bạn có chắc muốn vô hiệu hóa phòng ban <strong>{dept.name}</strong>?
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
