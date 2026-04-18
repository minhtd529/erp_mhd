'use client';
import * as React from 'react';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { commissionService, type RecordStatus } from '@/services/commissions';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatCurrency, formatDate } from '@/lib/utils';
import { CheckCircle, DollarSign, User } from 'lucide-react';

const STATUS_LABELS: Record<RecordStatus, string> = {
  accrued: 'Đã tích lũy', approved: 'Đã duyệt', paid: 'Đã chi', clawback: 'Hoàn lại',
};
const STATUS_VARIANTS: Record<RecordStatus, 'warning' | 'default' | 'success' | 'danger'> = {
  accrued: 'warning', approved: 'default', paid: 'success', clawback: 'danger',
};

export default function CommissionsPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const [statusFilter, setStatusFilter] = React.useState('accrued');

  const { data, isLoading } = useQuery({
    queryKey: ['commission-records', page, statusFilter],
    queryFn: () => commissionService.records.list({ page, size: 20, status: statusFilter || undefined }),
  });

  const approveMut = useMutation({
    mutationFn: (id: string) => commissionService.records.approve(id),
    onSuccess: () => { toast('Đã duyệt hoa hồng', 'success'); qc.invalidateQueries({ queryKey: ['commission-records'] }); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const payMut = useMutation({
    mutationFn: (id: string) => commissionService.records.markPaid(id),
    onSuccess: () => { toast('Đã đánh dấu đã chi', 'success'); qc.invalidateQueries({ queryKey: ['commission-records'] }); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-text-primary">Quản lý hoa hồng</h2>
          <p className="text-sm text-text-secondary">Duyệt và chi hoa hồng nhân viên</p>
        </div>
        <Link href="/commissions/my">
          <Button variant="outline"><User className="w-4 h-4" />Hoa hồng của tôi</Button>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Danh sách hoa hồng</CardTitle>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-44"><SelectValue placeholder="Tất cả" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="">Tất cả</SelectItem>
                {Object.entries(STATUS_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nhân viên</TableHead>
                  <TableHead>Hợp đồng</TableHead>
                  <TableHead>Số tiền</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead>Ngày tích lũy</TableHead>
                  <TableHead className="w-24">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((rec) => (
                  <TableRow key={rec.id}>
                    <TableCell className="text-sm font-medium">{rec.salesperson_name ?? '-'}</TableCell>
                    <TableCell className="text-xs text-text-secondary">{rec.engagement_title ?? '-'}</TableCell>
                    <TableCell className={`font-semibold ${rec.amount < 0 ? 'text-danger' : ''}`}>
                      {formatCurrency(rec.amount)}
                    </TableCell>
                    <TableCell><Badge variant={STATUS_VARIANTS[rec.status]}>{STATUS_LABELS[rec.status]}</Badge></TableCell>
                    <TableCell className="text-xs">{formatDate(rec.accrued_at)}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {rec.status === 'accrued' && (
                          <Button variant="ghost" size="icon" title="Duyệt" onClick={() => approveMut.mutate(rec.id)}>
                            <CheckCircle className="w-3.5 h-3.5 text-success" />
                          </Button>
                        )}
                        {rec.status === 'approved' && (
                          <Button variant="ghost" size="icon" title="Đánh dấu đã chi" onClick={() => payMut.mutate(rec.id)}>
                            <DollarSign className="w-3.5 h-3.5 text-action" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={6} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
      {data && <Pagination page={page} totalPages={data.total_pages} onPageChange={setPage} />}
    </div>
  );
}
