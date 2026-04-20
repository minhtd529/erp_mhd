'use client';
import * as React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { invoiceService, type InvoiceStatus } from '@/services/billing';
import { toast } from '@/hooks/use-toast';
import { getErrorMessage, formatDate, formatCurrency } from '@/lib/utils';
import { Download, Send, CheckCircle, XCircle, FileCheck } from 'lucide-react';

const STATUS_LABELS: Record<InvoiceStatus, string> = {
  DRAFT: 'Nháp', SENT: 'Đã gửi', CONFIRMED: 'Đã xác nhận',
  ISSUED: 'Đã xuất', PAID: 'Đã thanh toán', CANCELLED: 'Đã hủy',
};
const STATUS_VARIANTS: Record<InvoiceStatus, 'ghost' | 'warning' | 'default' | 'success' | 'secondary' | 'danger'> = {
  DRAFT: 'ghost', SENT: 'warning', CONFIRMED: 'default',
  ISSUED: 'secondary', PAID: 'success', CANCELLED: 'danger',
};

export default function InvoicesPage() {
  const qc = useQueryClient();
  const [page, setPage] = React.useState(1);
  const [statusFilter, setStatusFilter] = React.useState<string>('all');

  const { data, isLoading } = useQuery({
    queryKey: ['invoices', page, statusFilter],
    queryFn: () => invoiceService.list({ page, size: 20, status: statusFilter === 'all' ? undefined : statusFilter as InvoiceStatus }),
  });

  const transMut = useMutation({
    mutationFn: ({ id, action }: { id: string; action: string }) => invoiceService.transition(id, action),
    onSuccess: () => { toast('Cập nhật thành công', 'success'); qc.invalidateQueries({ queryKey: ['invoices'] }); },
    onError: (err) => toast(getErrorMessage(err), 'error'),
  });

  const handleExport = async () => {
    try {
      const blob = await invoiceService.exportCSV();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url; a.download = 'invoices.csv'; a.click();
      URL.revokeObjectURL(url);
    } catch { toast('Không thể xuất file', 'error'); }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-48"><SelectValue placeholder="Tất cả trạng thái" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Tất cả</SelectItem>
            {Object.entries(STATUS_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
          </SelectContent>
        </Select>
        <Button variant="outline" onClick={handleExport}><Download className="w-4 h-4" />Xuất CSV</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? <PageSpinner /> : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Số HĐ</TableHead>
                  <TableHead>Khách hàng</TableHead>
                  <TableHead>Tổng tiền</TableHead>
                  <TableHead>Ngày xuất</TableHead>
                  <TableHead>Hạn TT</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead className="w-28">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((inv) => (
                  <TableRow key={inv.id}>
                    <TableCell className="font-mono text-xs font-medium">{inv.number}</TableCell>
                    <TableCell className="text-sm">{inv.client_name ?? '-'}</TableCell>
                    <TableCell className="font-medium">{formatCurrency(inv.total_amount)}</TableCell>
                    <TableCell className="text-xs">{inv.issued_date ? formatDate(inv.issued_date) : '-'}</TableCell>
                    <TableCell className="text-xs">{inv.due_date ? formatDate(inv.due_date) : '-'}</TableCell>
                    <TableCell>
                      <Badge variant={STATUS_VARIANTS[inv.status]}>{STATUS_LABELS[inv.status]}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {inv.status === 'DRAFT' && (
                          <Button variant="ghost" size="icon" title="Gửi" onClick={() => transMut.mutate({ id: inv.id, action: 'send' })}>
                            <Send className="w-3.5 h-3.5" />
                          </Button>
                        )}
                        {inv.status === 'SENT' && (
                          <>
                            <Button variant="ghost" size="icon" title="Xác nhận" onClick={() => transMut.mutate({ id: inv.id, action: 'confirm' })}>
                              <CheckCircle className="w-3.5 h-3.5 text-success" />
                            </Button>
                            <Button variant="ghost" size="icon" title="Hủy" onClick={() => transMut.mutate({ id: inv.id, action: 'cancel' })}>
                              <XCircle className="w-3.5 h-3.5 text-danger" />
                            </Button>
                          </>
                        )}
                        {inv.status === 'CONFIRMED' && (
                          <Button variant="ghost" size="icon" title="Xuất hóa đơn" onClick={() => transMut.mutate({ id: inv.id, action: 'issue' })}>
                            <FileCheck className="w-3.5 h-3.5 text-action" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={7} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
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
