'use client';
import * as React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { PageSpinner } from '@/components/ui/spinner';
import { UpdateEmployeeForm } from '@/components/hrm/employee-form';
import { DependentSection } from '@/components/hrm/dependent-section';
import { ContractSection } from '@/components/hrm/contract-section';
import { SensitiveModal } from '@/components/hrm/sensitive-modal';
import { SalaryHistorySection } from '@/components/hrm/salary-history-section';
import { employeeService, type UpdateEmployeeRequest } from '@/services/hrm/employee';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { ArrowLeft, Pencil, ShieldAlert } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'HR_MANAGER'];
const SENSITIVE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO', 'HR_MANAGER'];
const SALARY_READ_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO', 'HR_MANAGER'];
const SALARY_WRITE_ROLES = ['SUPER_ADMIN', 'CEO', 'HR_MANAGER'];

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  ACTIVE: 'default', INACTIVE: 'secondary', ON_LEAVE: 'outline', TERMINATED: 'secondary',
};

const STATUS_LABELS: Record<string, string> = {
  ACTIVE: 'Đang làm', INACTIVE: 'Không HĐ', ON_LEAVE: 'Nghỉ phép', TERMINATED: 'Đã nghỉ',
};

type TabKey = 'info' | 'dependents' | 'contracts' | 'salary';

function InfoRow({ label, value }: { label: string; value?: string | number | null }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-text-secondary">{label}</span>
      <span className="text-sm text-text-primary">{value ?? '–'}</span>
    </div>
  );
}

function TabButton({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? 'border-action text-action'
          : 'border-transparent text-text-secondary hover:text-text-primary'
      }`}
    >
      {children}
    </button>
  );
}

export default function EmployeeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const qc = useQueryClient();
  const { user } = useAuthStore();

  const [tab, setTab] = React.useState<TabKey>('info');
  const [editOpen, setEditOpen] = React.useState(false);
  const [sensitiveOpen, setSensitiveOpen] = React.useState(false);

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));
  const canSeeSensitive = SENSITIVE_ROLES.some(r => user?.roles?.includes(r));
  const canReadSalary = SALARY_READ_ROLES.some(r => user?.roles?.includes(r));
  const canWriteSalary = SALARY_WRITE_ROLES.some(r => user?.roles?.includes(r));
  const canWriteDependent = [...WRITE_ROLES, 'HR_STAFF'].some(r => user?.roles?.includes(r));

  const { data: emp, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'employees', id],
    queryFn: () => employeeService.getById(id),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateEmployeeRequest) => employeeService.update(id, data),
    onSuccess: () => {
      toast('Cập nhật nhân viên thành công', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees', id] });
      setEditOpen(false);
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError || !emp) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không tìm thấy nhân viên.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 max-w-4xl">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/employees')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2 flex-wrap">
            <h1 className="text-xl font-semibold text-text-primary">{emp.full_name}</h1>
            <Badge variant={STATUS_VARIANTS[emp.status] ?? 'outline'}>
              {STATUS_LABELS[emp.status] ?? emp.status}
            </Badge>
            <Badge variant="outline">{emp.grade}</Badge>
          </div>
          <p className="text-xs text-text-secondary font-mono mt-0.5">{emp.email}</p>
        </div>
        <div className="flex gap-2">
          {canSeeSensitive && (
            <Button variant="outline" className="text-danger border-danger hover:bg-danger/5" onClick={() => setSensitiveOpen(true)}>
              <ShieldAlert className="w-4 h-4" />PII
            </Button>
          )}
          {canWrite && (
            <Button variant="outline" onClick={() => setEditOpen(true)}>
              <Pencil className="w-4 h-4" />Chỉnh sửa
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b flex gap-0">
        <TabButton active={tab === 'info'} onClick={() => setTab('info')}>Thông tin</TabButton>
        <TabButton active={tab === 'dependents'} onClick={() => setTab('dependents')}>Người phụ thuộc</TabButton>
        <TabButton active={tab === 'contracts'} onClick={() => setTab('contracts')}>Hợp đồng</TabButton>
        {canReadSalary && (
          <TabButton active={tab === 'salary'} onClick={() => setTab('salary')}>Lịch sử lương</TabButton>
        )}
      </div>

      {/* Tab: Thông tin */}
      {tab === 'info' && (
        <Card>
          <CardContent className="p-5">
            <div className="grid grid-cols-2 gap-4">
              <InfoRow label="Mã nhân viên" value={emp.employee_code} />
              <InfoRow label="Tên hiển thị" value={emp.display_name} />
              <InfoRow label="Điện thoại" value={emp.phone} />
              <InfoRow label="Ngày sinh" value={emp.date_of_birth} />
              <InfoRow label="Giới tính" value={emp.gender} />
              <InfoRow label="Quốc tịch" value={emp.nationality} />
              <InfoRow label="Chức vụ" value={emp.position_title} />
              <InfoRow label="Địa điểm làm việc" value={emp.work_location} />
              <InfoRow label="Ngày bắt đầu" value={emp.hired_date} />
              <InfoRow label="Ngày kết thúc thử việc" value={emp.probation_end_date} />
              <InfoRow label="Email cá nhân" value={emp.personal_email} />
              <InfoRow label="ĐT cá nhân" value={emp.personal_phone} />
              <InfoRow label="ĐT công ty" value={emp.work_phone} />
              <InfoRow label="Trình độ học vấn" value={emp.education_level} />
              <InfoRow label="Chuyên ngành" value={emp.education_major} />
              <InfoRow label="Trường học" value={emp.education_school} />
              <InfoRow label="Loại hoa hồng" value={emp.commission_type} />
              <InfoRow label="Số CPA" value={emp.vn_cpa_number} />
              <InfoRow label="Ngày cập nhật" value={formatDate(emp.updated_at)} />
              <InfoRow label="Ngày tạo" value={formatDate(emp.created_at)} />
              <div className="col-span-2">
                <InfoRow label="Địa chỉ hiện tại" value={emp.current_address} />
              </div>
              <div className="col-span-2">
                <InfoRow label="Địa chỉ thường trú" value={emp.permanent_address} />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tab: Người phụ thuộc */}
      {tab === 'dependents' && (
        <Card>
          <CardContent className="p-5">
            <DependentSection employeeId={id} canWrite={canWriteDependent} />
          </CardContent>
        </Card>
      )}

      {/* Tab: Hợp đồng */}
      {tab === 'contracts' && (
        <Card>
          <CardContent className="p-5">
            <ContractSection employeeId={id} canWrite={canWrite} />
          </CardContent>
        </Card>
      )}

      {/* Tab: Lịch sử lương */}
      {tab === 'salary' && canReadSalary && (
        <Card>
          <CardContent className="p-5">
            <SalaryHistorySection
              employeeId={id}
              canRead={canReadSalary}
              canWrite={canWriteSalary}
            />
          </CardContent>
        </Card>
      )}

      {/* Edit dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle>Chỉnh sửa nhân viên</DialogTitle></DialogHeader>
          <UpdateEmployeeForm
            formId="emp-edit-form"
            initial={emp}
            onSubmit={(data) => updateMutation.mutate(data)}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Hủy</Button>
            <Button type="submit" form="emp-edit-form" loading={updateMutation.isPending}>Lưu</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Sensitive PII modal */}
      {canSeeSensitive && (
        <SensitiveModal
          open={sensitiveOpen}
          onOpenChange={setSensitiveOpen}
          employeeId={id}
          employeeName={emp.full_name}
        />
      )}
    </div>
  );
}
