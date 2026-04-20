'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { paymentService, type PaymentStatus } from '@/services/billing';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate, formatCurrency } from '@/lib/utils';
import { CheckCircle, AlertCircle, RotateCcw } from 'lucide-react';

const STATUS_LABELS: Record<PaymentStatus, string> = {
  RECORDED: 'Đã ghi nhận', CLEARED: 'Đã thanh toán', DISPUTED: 'Đang tranh chấp', REVERSED: 'Đã đảo',
};
const STATUS_VARIANTS: Record<PaymentStatus, 'warning' | 'success' | 'danger' | 'secondary'> = {
  RECORDED: 'warning', CLEARED: 'success', DISPUTED: 'danger', REVERSED: 'secondary',
};

export default function PaymentsPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const { data, isLoading, isError } = useQuery({
    queryKey: ['payments', page],
    queryFn: () => paymentService.list({ page }),
  });

  const clearMut = useMutation({ mutationFn: (id: string) => paymentService.clear(id), onSuccess: () => { toast('Đã xác nhận thanh toán', 'success'); qc.invalidateQueries({ queryKey: ['payments'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });
  const disputeMut = useMutation({ mutationFn: (id: string) => paymentService.dispute(id), onSuccess: () => { toast('Đã ghi nhận tranh chấp', 'success'); qc.invalidateQueries({ queryKey: ['payments'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });
  const reverseMut = useMutation({ mutationFn: (id: string) => paymentService.reverse(id), onSuccess: () => { toast('Đã đảo thanh toán', 'success'); qc.invalidateQueries({ queryKey: ['payments'] }); }, onError: (err) => toast(getErrorMessage(err), 'error') });

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : isError ? (
              <div className="text-center text-text-secondary py-8 text-sm">API chưa sẵn sàng. Vui lòng thử lại sau.</div>
            ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Số HĐ</TableHead>
                  <TableHead>Số tiền</TableHead>
                  <TableHead>Phương thức</TableHead>
                  <TableHead>Ngày TT</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="w-28">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell className="font-mono text-xs">{p.invoice_id.slice(0, 8)}…</TableCell>
                    <TableCell className="font-medium">{formatCurrency(p.amount)}</TableCell>
                    <TableCell className="text-xs">{p.payment_method ?? '-'}</TableCell>
                    <TableCell className="text-xs">{p.payment_date ? formatDate(p.payment_date) : '-'}</TableCell>
                    <TableCell><Badge variant={STATUS_VARIANTS[p.status]}>{STATUS_LABELS[p.status]}</Badge></TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {p.status === 'RECORDED' && (
                          <>
                            <Button variant="ghost" size="icon" title="Xác nhận" onClick={() => clearMut.mutate(p.id)}>
                              <CheckCircle className="w-3.5 h-3.5 text-success" />
                            </Button>
                            <Button variant="ghost" size="icon" title="Tranh chấp" onClick={() => disputeMut.mutate(p.id)}>
                              <AlertCircle className="w-3.5 h-3.5 text-danger" />
                            </Button>
                          </>
                        )}
                        {p.status === 'CLEARED' && (
                          <Button variant="ghost" size="icon" title="Đảo giao dịch" onClick={() => { if (confirm('Đảo giao dịch này?')) reverseMut.mutate(p.id); }}>
                            <RotateCcw className="w-3.5 h-3.5 text-text-secondary" />
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
