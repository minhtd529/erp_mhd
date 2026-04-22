'use client';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { employeeService } from '@/services/hrm/employee';
import { provisioningService, type CreateProvisioningRequest } from '@/services/hrm/provisioning';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage } from '@/lib/utils';
import { MODULE_ROLES } from '@/lib/roles';
import { ArrowLeft, AlertTriangle } from 'lucide-react';

const ROLE_OPTIONS = [
  { value: 'HR_MANAGER',     label: 'Trưởng Nhân sự' },
  { value: 'HR_STAFF',       label: 'NV Nhân sự' },
  { value: 'HEAD_OF_BRANCH', label: 'Trưởng Chi nhánh' },
  { value: 'AUDIT_MANAGER',  label: 'QL Kiểm toán' },
  { value: 'SENIOR_AUDITOR', label: 'KTV Cao cấp' },
  { value: 'JUNIOR_AUDITOR', label: 'Kiểm toán viên' },
  { value: 'AUDIT_STAFF',    label: 'NV Kiểm toán' },
  { value: 'ACCOUNTANT',     label: 'Kế toán' },
  { value: 'FIRM_PARTNER',   label: 'Partner' },
  { value: 'CEO',            label: 'Tổng Giám đốc' },
];

function FieldLabel({ children, required }: { children: React.ReactNode; required?: boolean }) {
  return (
    <label className="text-sm font-medium text-text-primary">
      {children}{required && <span className="text-danger ml-0.5">*</span>}
    </label>
  );
}

export default function NewProvisioningPage() {
  const router = useRouter();
  const { user } = useAuthStore();

  const [employeeId, setEmployeeId] = React.useState('');
  const [requestedRole, setRequestedRole] = React.useState('');
  const [isEmergency, setIsEmergency] = React.useState(false);
  const [emergencyReason, setEmergencyReason] = React.useState('');
  const [notes, setNotes] = React.useState('');
  const [empSearch, setEmpSearch] = React.useState('');

  const canCreate = MODULE_ROLES.hrmProvisioningCreate.some(r => user?.roles?.includes(r));

  React.useEffect(() => {
    if (!canCreate) router.replace('/admin/hrm/provisioning');
  }, [canCreate, router]);

  const { data: empData } = useQuery({
    queryKey: ['hrm', 'employees', 1, empSearch],
    queryFn: () => employeeService.list({ page: 1, size: 50, q: empSearch || undefined }),
    staleTime: 10_000,
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateProvisioningRequest) => provisioningService.create(data),
    onSuccess: (req) => {
      toast('Tạo yêu cầu cấp tài khoản thành công', 'success');
      router.push(`/admin/hrm/provisioning/${req.id}`);
    },
    onError: (err) => {
      const msg = getErrorMessage(err);
      if (msg.includes('DUPLICATE_PENDING')) {
        toast('Nhân viên đã có yêu cầu đang chờ duyệt', 'error');
      } else {
        toast(msg, 'error');
      }
    },
  });

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!employeeId || !requestedRole) return;
    createMutation.mutate({
      employee_id: employeeId,
      requested_role: requestedRole,
      is_emergency: isEmergency,
      emergency_reason: isEmergency ? emergencyReason : undefined,
      notes: notes || undefined,
    });
  }

  const employees = empData?.data ?? [];

  return (
    <div className="flex flex-col gap-4 max-w-2xl">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => router.push('/admin/hrm/provisioning')}>
          <ArrowLeft className="w-4 h-4" />
        </Button>
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Tạo yêu cầu cấp tài khoản</h1>
          <p className="text-sm text-text-secondary">Yêu cầu cấp quyền truy cập hệ thống cho nhân viên</p>
        </div>
      </div>

      <Card>
        <CardContent className="p-5">
          <form id="create-provisioning-form" onSubmit={handleSubmit} className="flex flex-col gap-5">
            <div className="flex flex-col gap-1.5">
              <FieldLabel required>Tìm nhân viên</FieldLabel>
              <input
                type="text"
                placeholder="Tìm theo tên hoặc email..."
                value={empSearch}
                onChange={e => setEmpSearch(e.target.value)}
                className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30"
              />
              {employees.length > 0 && (
                <select
                  value={employeeId}
                  onChange={e => setEmployeeId(e.target.value)}
                  className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30"
                  size={Math.min(employees.length, 6)}
                >
                  <option value="">-- Chọn nhân viên --</option>
                  {employees.map(emp => (
                    <option key={emp.id} value={emp.id}>
                      {emp.full_name} — {emp.email}
                    </option>
                  ))}
                </select>
              )}
              {empSearch && employees.length === 0 && (
                <p className="text-xs text-text-secondary">Không tìm thấy nhân viên.</p>
              )}
            </div>

            <div className="flex flex-col gap-1.5">
              <FieldLabel required>Vai trò yêu cầu</FieldLabel>
              <select
                value={requestedRole}
                onChange={e => setRequestedRole(e.target.value)}
                className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30"
              >
                <option value="">-- Chọn vai trò --</option>
                {ROLE_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>

            <div className="flex flex-col gap-1.5">
              <FieldLabel>Ghi chú</FieldLabel>
              <textarea
                value={notes}
                onChange={e => setNotes(e.target.value)}
                rows={3}
                placeholder="Ghi chú thêm (tùy chọn)..."
                className="rounded-md border border-border bg-surface-paper px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30 resize-none"
              />
            </div>

            <div className="flex flex-col gap-2 rounded-md border border-border p-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isEmergency}
                  onChange={e => setIsEmergency(e.target.checked)}
                  className="w-4 h-4 accent-action"
                />
                <span className="text-sm font-medium text-text-primary flex items-center gap-1.5">
                  <AlertTriangle className="w-4 h-4 text-warning" />
                  Yêu cầu khẩn cấp
                </span>
              </label>
              {isEmergency && (
                <div className="flex flex-col gap-1.5 mt-1">
                  <FieldLabel required>Lý do khẩn cấp</FieldLabel>
                  <textarea
                    value={emergencyReason}
                    onChange={e => setEmergencyReason(e.target.value)}
                    rows={2}
                    placeholder="Mô tả lý do khẩn cấp..."
                    className="rounded-md border border-border bg-surface-paper px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30 resize-none"
                    required
                  />
                </div>
              )}
            </div>
          </form>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={() => router.push('/admin/hrm/provisioning')}>Hủy</Button>
        <Button
          type="submit"
          form="create-provisioning-form"
          loading={createMutation.isPending}
          disabled={!employeeId || !requestedRole || (isEmergency && !emergencyReason)}
        >
          Tạo yêu cầu
        </Button>
      </div>
    </div>
  );
}
