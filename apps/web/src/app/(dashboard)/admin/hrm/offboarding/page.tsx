'use client';
import * as React from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { offboardingService } from '@/services/hrm/provisioning';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { formatDate } from '@/lib/utils';
import { MODULE_ROLES } from '@/lib/roles';
import { ClipboardList, Eye, Plus } from 'lucide-react';

const STATUS_LABELS: Record<string, string> = {
  IN_PROGRESS: 'Đang xử lý',
  COMPLETED:   'Hoàn thành',
  CANCELLED:   'Đã hủy',
};

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  IN_PROGRESS: 'outline',
  COMPLETED:   'default',
  CANCELLED:   'secondary',
};

const TYPE_LABELS: Record<string, string> = {
  ONBOARDING:  'Onboarding',
  OFFBOARDING: 'Offboarding',
};

const STATUS_FILTER_OPTIONS = [
  { value: '',            label: 'Tất cả trạng thái' },
  { value: 'IN_PROGRESS', label: 'Đang xử lý' },
  { value: 'COMPLETED',   label: 'Hoàn thành' },
  { value: 'CANCELLED',   label: 'Đã hủy' },
];

export default function OffboardingListPage() {
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [status, setStatus] = React.useState('');

  const canCreate = MODULE_ROLES.hrmOffboardingCreate.some(r => user?.roles?.includes(r));

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'offboarding', page, status],
    queryFn: () => offboardingService.list({ page, size: 20, status: status || undefined }),
  });

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải danh sách.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const items = data?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Offboarding</h1>
          <p className="text-sm text-text-secondary">Quản lý checklist offboarding nhân viên</p>
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
              <Link href="/admin/hrm/offboarding/new">
                <Plus className="w-4 h-4" />Tạo checklist
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
                <TableHead>Loại</TableHead>
                <TableHead>Trạng thái</TableHead>
                <TableHead>Ngày mục tiêu</TableHead>
                <TableHead>Ngày tạo</TableHead>
                <TableHead className="w-20"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-mono text-xs">{item.employee_id.slice(0, 8)}…</TableCell>
                  <TableCell>
                    <Badge variant="outline">{TYPE_LABELS[item.checklist_type] ?? item.checklist_type}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANTS[item.status] ?? 'outline'}>
                      {STATUS_LABELS[item.status] ?? item.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-text-secondary">{item.target_date ? formatDate(item.target_date) : '–'}</TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(item.created_at)}</TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" asChild title="Xem chi tiết">
                      <Link href={`/admin/hrm/offboarding/${item.id}`}>
                        <Eye className="w-3.5 h-3.5" />
                      </Link>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-16">
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <ClipboardList className="w-10 h-10 opacity-30" />
                      <p>Không có checklist nào.</p>
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
