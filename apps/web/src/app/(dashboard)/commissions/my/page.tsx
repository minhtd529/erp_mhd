'use client';
import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { PageSpinner } from '@/components/ui/spinner';
import { Pagination } from '@/components/shared/pagination';
import { commissionService, type RecordStatus } from '@/services/commissions';
import { formatCurrency, formatDate } from '@/lib/utils';
import { TrendingUp, Clock, DollarSign, PauseCircle } from 'lucide-react';

const STATUS_LABELS: Record<RecordStatus, string> = {
  accrued: 'Đã tích lũy', approved: 'Đã duyệt', paid: 'Đã chi', clawback: 'Hoàn lại',
};
const STATUS_VARIANTS: Record<RecordStatus, 'warning' | 'default' | 'success' | 'danger'> = {
  accrued: 'warning', approved: 'default', paid: 'success', clawback: 'danger',
};

function SummaryCard({ title, value, icon: Icon, color }: { title: string; value: number; icon: React.ComponentType<{className?: string}>; color: string }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-4 pt-5">
        <div className={`w-10 h-10 rounded-card flex items-center justify-center flex-shrink-0 ${color}`}>
          <Icon className="w-5 h-5" />
        </div>
        <div>
          <p className="text-xs text-text-secondary">{title}</p>
          <p className="text-lg font-bold text-text-primary">{formatCurrency(value)}</p>
        </div>
      </CardContent>
    </Card>
  );
}

export default function MyCommissionsPage() {
  const [page, setPage] = React.useState(1);
  const [statusFilter, setStatusFilter] = React.useState('all');

  const { data: summary, isLoading: summaryLoading } = useQuery({
    queryKey: ['my-commissions-summary'],
    queryFn: commissionService.me.summary,
  });

  const { data, isLoading } = useQuery({
    queryKey: ['my-commissions', page, statusFilter],
    queryFn: () => commissionService.me.list({ page, size: 20, status: statusFilter === 'all' ? undefined : statusFilter as RecordStatus }),
  });

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">Hoa hồng của tôi</h2>
        <p className="text-sm text-text-secondary">Theo dõi và quản lý hoa hồng cá nhân</p>
      </div>

      {summaryLoading ? <PageSpinner /> : summary && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <SummaryCard
            title="Cả năm (YTD)"
            value={summary.ytd_accrued}
            icon={TrendingUp}
            color="bg-primary/10 text-primary"
          />
          <SummaryCard
            title="Tháng này"
            value={summary.month_accrued}
            icon={DollarSign}
            color="bg-success/10 text-success"
          />
          <SummaryCard
            title="Chờ duyệt"
            value={summary.pending_approval}
            icon={Clock}
            color="bg-amber-100 text-amber-700"
          />
          <SummaryCard
            title="Đang giữ"
            value={summary.on_hold}
            icon={PauseCircle}
            color="bg-gray-100 text-text-secondary"
          />
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-2 gap-4">
        <Card>
          <CardHeader><CardTitle className="text-sm">Đã thanh toán</CardTitle></CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-success">{summary ? formatCurrency(summary.ytd_paid) : '-'}</p>
            <p className="text-xs text-text-secondary mt-1">Tổng đã nhận từ đầu năm</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-sm">Tháng này đã nhận</CardTitle></CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-action">{summary ? formatCurrency(summary.month_paid) : '-'}</p>
            <p className="text-xs text-text-secondary mt-1">Đã được chi trả tháng hiện tại</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Lịch sử hoa hồng</CardTitle>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-40"><SelectValue placeholder="Tất cả" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Tất cả</SelectItem>
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
                  <TableHead>Hợp đồng</TableHead>
                  <TableHead>Số tiền</TableHead>
                  <TableHead>Trạng thái</TableHead>
                  <TableHead>Ngày tích lũy</TableHead>
                  <TableHead>Ngày chi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((rec) => (
                  <TableRow key={rec.id}>
                    <TableCell className="text-sm">{rec.engagement_title ?? rec.engagement_commission_id.slice(0, 8)}</TableCell>
                    <TableCell className={`font-semibold ${rec.amount < 0 ? 'text-danger' : 'text-text-primary'}`}>
                      {formatCurrency(rec.amount)}
                    </TableCell>
                    <TableCell><Badge variant={STATUS_VARIANTS[rec.status]}>{STATUS_LABELS[rec.status]}</Badge></TableCell>
                    <TableCell className="text-xs">{formatDate(rec.accrued_at)}</TableCell>
                    <TableCell className="text-xs">{rec.paid_at ? formatDate(rec.paid_at) : '-'}</TableCell>
                  </TableRow>
                ))}
                {data?.data.length === 0 && (
                  <TableRow><TableCell colSpan={5} className="text-center text-text-secondary py-8">Không có dữ liệu</TableCell></TableRow>
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
