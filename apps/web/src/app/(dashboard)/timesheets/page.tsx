'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { timesheetService, type TimesheetStatus } from '@/services/timesheets';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate } from '@/lib/utils';
import { CheckCircle, XCircle, Send, Lock } from 'lucide-react';

const STATUS_LABELS: Record<TimesheetStatus, string> = {
  OPEN: 'Đang mở', SUBMITTED: 'Chờ duyệt', APPROVED: 'Đã duyệt',
  REJECTED: 'Từ chối', LOCKED: 'Đã khóa',
};
const STATUS_VARIANTS: Record<TimesheetStatus, 'ghost' | 'warning' | 'success' | 'danger' | 'secondary'> = {
  OPEN: 'ghost', SUBMITTED: 'warning', APPROVED: 'success', REJECTED: 'danger', LOCKED: 'secondary',
};

export default function TimesheetsPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['timesheets', page],
    queryFn: () => timesheetService.list({ page, size: 20 }),
  });

  const mutate = (fn: () => Promise<unknown>, successMsg: string) => ({
    mutationFn: fn,
    onSuccess: () => { toast(successMsg, 'success'); qc.invalidateQueries({ queryKey: ['timesheets'] }); },
    onError: (err: unknown) => toast(getErrorMessage(err), 'error'),
  });

  const submitMut = useMutation({ mutationFn: (id: string) => timesheetService.submit(id), onSuccess: () => { toast('Đã gửi duyệt', 'success'); qc.invalidateQueries({ queryKey: ['timesheets'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });
  const approveMut = useMutation({ mutationFn: (id: string) => timesheetService.approve(id), onSuccess: () => { toast('Đã duyệt', 'success'); qc.invalidateQueries({ queryKey: ['timesheets'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });
  const rejectMut = useMutation({ mutationFn: (id: string) => timesheetService.reject(id), onSuccess: () => { toast('Đã từ chối', 'success'); qc.invalidateQueries({ queryKey: ['timesheets'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });
  const lockMut = useMutation({ mutationFn: (id: string) => timesheetService.lock(id), onSuccess: () => { toast('Đã khóa', 'success'); qc.invalidateQueries({ queryKey: ['timesheets'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nhân viên</TableHead>
                  <TableHead>Kỳ</TableHead>
                  <TableHead>Tổng giờ</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead>Ngày gửi</TableHead>
                  <TableHead className="w-36">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((ts) => (
                  <TableRow key={ts.id}>
                    <TableCell className="font-medium text-sm">{ts.employee_name ?? ts.employee_id}</TableCell>
                    <TableCell className="text-xs">
                      {formatDate(ts.period_start)} – {formatDate(ts.period_end)}
                    </TableCell>
                    <TableCell>{ts.total_hours ? `${ts.total_hours}h` : '-'}</TableCell>
                    <TableCell>
                      <Badge variant={STATUS_VARIANTS[ts.status]}>{STATUS_LABELS[ts.status]}</Badge>
                    </TableCell>
                    <TableCell className="text-xs">{ts.submitted_at ? formatDate(ts.submitted_at) : '-'}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {ts.status === 'OPEN' && (
                          <Button variant="ghost" size="icon" title="Gửi duyệt" onClick={() => submitMut.mutate(ts.id)}>
                            <Send className="w-3.5 h-3.5" />
                          </Button>
                        )}
                        {ts.status === 'SUBMITTED' && (
                          <>
                            <Button variant="ghost" size="icon" title="Duyệt" onClick={() => approveMut.mutate(ts.id)}>
                              <CheckCircle className="w-3.5 h-3.5 text-success" />
                            </Button>
                            <Button variant="ghost" size="icon" title="Từ chối" onClick={() => rejectMut.mutate(ts.id)}>
                              <XCircle className="w-3.5 h-3.5 text-danger" />
                            </Button>
                          </>
                        )}
                        {ts.status === 'APPROVED' && (
                          <Button variant="ghost" size="icon" title="Khóa" onClick={() => lockMut.mutate(ts.id)}>
                            <Lock className="w-3.5 h-3.5 text-text-secondary" />
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
