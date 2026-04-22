'use client';
import * as React from 'react';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { provisioningService, type ProvisioningRequest } from '@/services/hrm/provisioning';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { MODULE_ROLES } from '@/lib/roles';
import { UserPlus, Eye, Plus } from 'lucide-react';

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
  PARTNER:        'Partner',
  AUDIT_MANAGER:  'QL Kiểm toán',
  SENIOR_AUDITOR: 'KTV Cao cấp',
  JUNIOR_AUDITOR: 'Kiểm toán viên',
  AUDIT_STAFF:    'NV Kiểm toán',
  ACCOUNTANT:     'Kế toán',
};

const STATUS_FILTER_OPTIONS = [
  { value: '', label: 'Tất cả trạng thái' },
  { value: 'PENDING', label: 'Chờ duyệt' },
  { value: 'APPROVED', label: 'Đã duyệt' },
  { value: 'REJECTED', label: 'Từ chối' },
  { value: 'EXECUTED', label: 'Đã thực thi' },
  { value: 'CANCELLED', label: 'Đã hủy' },
];

export default function ProvisioningListPage() {
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [status, setStatus] = React.useState('');

  const canCreate = MODULE_ROLES.hrmProvisioningCreate.some(r => user?.roles?.includes(r));

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'provisioning', page, status],
    queryFn: () => provisioningService.list({ page, size: 20, status: status || undefined }),
  });

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải danh sách yêu cầu.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const items = data?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Cấp tài khoản hệ thống</h1>
          <p className="text-sm text-text-secondary">Quản lý yêu cầu cấp quyền truy cập cho nhân viên</p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={status}
            onChange={e => { setStatus(e.target.value); setPage(1); }}
            className="h-9 rounded-md border border-border bg-surface-paper px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-action/30"
          >
            {STATUS_FILTER_OPTIONS.map(opt => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
          {canCreate && (
            <Button asChild>
              <Link href="/admin/hrm/provisioning/new">
                <Plus className="w-4 h-4" />Tạo yêu cầu
              </Link>
            </Button>
          )}
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nhân viên</TableHead>
                <TableHead>Vai trò yêu cầu</TableHead>
                <TableHead>Trạng thái</TableHead>
                <TableHead>Khẩn cấp</TableHead>
                <TableHead>Hết hạn</TableHead>
                <TableHead>Ngày tạo</TableHead>
                <TableHead className="w-20"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((req) => (
                <TableRow key={req.id}>
                  <TableCell className="font-mono text-xs">{req.employee_id.slice(0, 8)}…</TableCell>
                  <TableCell>
                    <Badge variant="outline">{ROLE_LABELS[req.requested_role] ?? req.requested_role}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANTS[req.status] ?? 'outline'}>
                      {STATUS_LABELS[req.status] ?? req.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {req.is_emergency
                      ? <Badge variant="secondary" className="text-danger border-danger/30">Khẩn</Badge>
                      : <span className="text-xs text-text-secondary">–</span>
                    }
                  </TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(req.expires_at)}</TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(req.created_at)}</TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" asChild title="Xem chi tiết">
                      <Link href={`/admin/hrm/provisioning/${req.id}`}>
                        <Eye className="w-3.5 h-3.5" />
                      </Link>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-16">
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <UserPlus className="w-10 h-10 opacity-30" />
                      <p>Không có yêu cầu nào.</p>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {data && data.total_pages > 1 && (
        <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />
      )}
    </div>
  );
}
