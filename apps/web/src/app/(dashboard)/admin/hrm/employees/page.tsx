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
import { employeeService } from '@/services/hrm/employee';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Users, Pencil, Trash2, Plus } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'HR_MANAGER'];

const GRADE_LABELS: Record<string, string> = {
  EXECUTIVE: 'Executive', PARTNER: 'Partner', DIRECTOR: 'Director',
  MANAGER: 'Manager', SENIOR: 'Senior', JUNIOR: 'Junior', INTERN: 'Intern', SUPPORT: 'Support',
};

const STATUS_VARIANTS: Record<string, 'default' | 'secondary' | 'outline'> = {
  ACTIVE: 'default', INACTIVE: 'secondary', ON_LEAVE: 'outline', TERMINATED: 'secondary',
};

const STATUS_LABELS: Record<string, string> = {
  ACTIVE: 'Đang làm', INACTIVE: 'Không HĐ', ON_LEAVE: 'Nghỉ phép', TERMINATED: 'Đã nghỉ',
};

export default function EmployeesPage() {
  const qc = useQueryClient();
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'employees', page, q],
    queryFn: () => employeeService.list({ page, size: 20, q: q || undefined }),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => employeeService.delete(id),
    onSuccess: () => {
      toast('Đã xóa nhân viên', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'employees'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải danh sách nhân viên.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  const employees = data?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Nhân viên</h1>
          <p className="text-sm text-text-secondary">Quản lý danh sách nhân viên</p>
        </div>
        <div className="flex items-center gap-2">
          <SearchInput placeholder="Tìm theo tên, email..." className="w-72" onSearch={setQ} />
          {canWrite && (
            <Button asChild>
              <Link href="/admin/hrm/employees/new">
                <Plus className="w-4 h-4" />Thêm nhân viên
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
                <TableHead>Mã</TableHead>
                <TableHead>Họ và tên</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Cấp bậc</TableHead>
                <TableHead>Trạng thái</TableHead>
                <TableHead>Loại HĐ</TableHead>
                <TableHead>Ngày tạo</TableHead>
                <TableHead className="w-24"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {employees.map((emp) => (
                <TableRow key={emp.id}>
                  <TableCell className="font-mono text-xs">{emp.employee_code ?? '–'}</TableCell>
                  <TableCell>
                    <Link
                      href={`/admin/hrm/employees/${emp.id}`}
                      className="font-medium text-action hover:underline"
                    >
                      {emp.full_name}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm">{emp.email}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{GRADE_LABELS[emp.grade] ?? emp.grade}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={STATUS_VARIANTS[emp.status] ?? 'outline'}>
                      {STATUS_LABELS[emp.status] ?? emp.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm">{emp.employment_type}</TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(emp.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" asChild>
                        <Link href={`/admin/hrm/employees/${emp.id}`} title="Chi tiết">
                          <Pencil className="w-3.5 h-3.5" />
                        </Link>
                      </Button>
                      {canWrite && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-danger hover:text-danger"
                          title="Xóa"
                          onClick={() => {
                            if (confirm(`Xóa nhân viên "${emp.full_name}"?`)) {
                              deleteMutation.mutate(emp.id);
                            }
                          }}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {employees.length === 0 && (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-16">
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <Users className="w-10 h-10 opacity-30" />
                      <p>Không tìm thấy nhân viên nào.</p>
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
