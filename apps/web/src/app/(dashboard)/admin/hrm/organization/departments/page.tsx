'use client';
import * as React from 'react';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { SearchInput } from '@/components/shared/search-input';
import { Pagination } from '@/components/shared/pagination';
import { departmentService } from '@/services/hrm/organization';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Layers, Pencil, PowerOff } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO'];

export default function DepartmentsPage() {
  const qc = useQueryClient();
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'departments', page, q],
    queryFn: () => departmentService.list({ page, size: 20, q: q || undefined }),
  });

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => departmentService.deactivate(id),
    onSuccess: () => {
      toast('Đã vô hiệu hóa phòng ban', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'departments'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải danh sách phòng ban.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const departments = data?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Phòng ban</h1>
          <p className="text-sm text-text-secondary">Quản lý danh sách phòng ban</p>
        </div>
        <SearchInput placeholder="Tìm theo tên, mã..." className="w-72" onSearch={setQ} />
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Mã</TableHead>
                <TableHead>Tên phòng ban</TableHead>
                <TableHead>Loại</TableHead>
                <TableHead>Trạng thái</TableHead>
                <TableHead>Ngày tạo</TableHead>
                <TableHead className="w-24"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {departments.map((dept) => (
                <TableRow key={dept.id}>
                  <TableCell className="font-mono text-xs">{dept.code}</TableCell>
                  <TableCell>
                    <Link href={`/admin/hrm/organization/departments/${dept.id}`} className="font-medium text-action hover:underline">
                      {dept.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant={dept.dept_type === 'CORE' ? 'default' : 'secondary'}>
                      {dept.dept_type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={dept.is_active ? 'default' : 'secondary'}>
                      {dept.is_active ? 'Hoạt động' : 'Ngưng'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(dept.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" asChild>
                        <Link href={`/admin/hrm/organization/departments/${dept.id}`} title="Chỉnh sửa">
                          <Pencil className="w-3.5 h-3.5" />
                        </Link>
                      </Button>
                      {canWrite && dept.is_active && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-danger hover:text-danger"
                          title="Vô hiệu hóa"
                          onClick={() => {
                            if (confirm(`Vô hiệu hóa phòng ban "${dept.name}"?`)) {
                              deactivateMutation.mutate(dept.id);
                            }
                          }}
                        >
                          <PowerOff className="w-3.5 h-3.5" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {departments.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-16">
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <Layers className="w-10 h-10 opacity-30" />
                      <p>Chưa có phòng ban nào.</p>
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
