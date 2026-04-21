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
import { branchService } from '@/services/hrm/organization';
import { useAuthStore } from '@/stores/auth';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { Building2, Pencil, PowerOff } from 'lucide-react';

const WRITE_ROLES = ['SUPER_ADMIN', 'CHAIRMAN', 'CEO'];
const HOB_ROLES = [...WRITE_ROLES, 'HEAD_OF_BRANCH'];

export default function BranchesPage() {
  const qc = useQueryClient();
  const { user } = useAuthStore();
  const [page, setPage] = React.useState(1);
  const [q, setQ] = React.useState('');

  const canWrite = WRITE_ROLES.some(r => user?.roles?.includes(r));
  const isHoB = user?.roles?.includes('HEAD_OF_BRANCH') && !canWrite;

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['hrm', 'branches', page, q],
    queryFn: () => branchService.list({ page, size: 20, q: q || undefined }),
  });

  const deactivateMutation = useMutation({
    mutationFn: (id: string) => branchService.deactivate(id),
    onSuccess: () => {
      toast('Đã vô hiệu hóa chi nhánh', 'success');
      qc.invalidateQueries({ queryKey: ['hrm', 'branches'] });
    },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  // HEAD_OF_BRANCH: filter to own branch only
  const branches = React.useMemo(() => {
    if (!data?.data) return [];
    if (isHoB && user?.branch_id) {
      return data.data.filter(b => b.id === user.branch_id);
    }
    return data.data;
  }, [data, isHoB, user?.branch_id]);

  if (isLoading) return <PageSpinner />;

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 py-16 text-text-secondary">
        <p>Không thể tải danh sách chi nhánh.</p>
        <Button variant="outline" onClick={() => refetch()}>Thử lại</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Chi nhánh</h1>
          <p className="text-sm text-text-secondary">Quản lý danh sách chi nhánh của công ty</p>
        </div>
        {!isHoB && (
          <SearchInput placeholder="Tìm theo tên, mã..." className="w-72" onSearch={setQ} />
        )}
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Mã</TableHead>
                <TableHead>Tên chi nhánh</TableHead>
                <TableHead>Thành phố</TableHead>
                <TableHead>Trụ sở</TableHead>
                <TableHead>Trạng thái</TableHead>
                <TableHead>Ngày tạo</TableHead>
                {HOB_ROLES.some(r => user?.roles?.includes(r)) && (
                  <TableHead className="w-24"></TableHead>
                )}
              </TableRow>
            </TableHeader>
            <TableBody>
              {branches.map((branch) => (
                <TableRow key={branch.id} className="cursor-pointer hover:bg-background">
                  <TableCell className="font-mono text-xs">{branch.code}</TableCell>
                  <TableCell>
                    <Link href={`/admin/hrm/organization/branches/${branch.id}`} className="font-medium text-action hover:underline">
                      {branch.name}
                    </Link>
                  </TableCell>
                  <TableCell className="text-sm">{branch.city ?? '–'}</TableCell>
                  <TableCell>
                    {branch.is_head_office && <Badge variant="outline" className="text-secondary border-secondary">Trụ sở</Badge>}
                  </TableCell>
                  <TableCell>
                    <Badge variant={branch.is_active ? 'default' : 'secondary'}>
                      {branch.is_active ? 'Hoạt động' : 'Ngưng'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-text-secondary">{formatDate(branch.created_at)}</TableCell>
                  {HOB_ROLES.some(r => user?.roles?.includes(r)) && (
                    <TableCell>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" asChild>
                          <Link href={`/admin/hrm/organization/branches/${branch.id}`} title="Chỉnh sửa">
                            <Pencil className="w-3.5 h-3.5" />
                          </Link>
                        </Button>
                        {canWrite && branch.is_active && (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="text-danger hover:text-danger"
                            title="Vô hiệu hóa"
                            onClick={() => {
                              if (confirm(`Vô hiệu hóa chi nhánh "${branch.name}"?`)) {
                                deactivateMutation.mutate(branch.id);
                              }
                            }}
                          >
                            <PowerOff className="w-3.5 h-3.5" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  )}
                </TableRow>
              ))}
              {branches.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-16">
                    <div className="flex flex-col items-center gap-2 text-text-secondary">
                      <Building2 className="w-10 h-10 opacity-30" />
                      <p>Chưa có chi nhánh nào. Liên hệ quản trị viên.</p>
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
