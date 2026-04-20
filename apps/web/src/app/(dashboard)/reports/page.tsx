'use client';
import * as React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { PageSpinner } from '@/components/ui/spinner';
import { reportService } from '@/services/reports';
import { useAuthStore } from '@/stores/auth';
import { formatCurrency } from '@/lib/utils';
import { BarChart3, TrendingUp, Users, Briefcase, DollarSign, AlertCircle } from 'lucide-react';

function KpiRow({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="flex items-center justify-between py-2.5 border-b border-border last:border-0">
      <span className="text-sm text-text-secondary">{label}</span>
      <div className="text-right">
        <span className="text-sm font-semibold text-text-primary">{value}</span>
        {sub && <p className="text-xs text-text-secondary">{sub}</p>}
      </div>
    </div>
  );
}

function ExecutiveReport() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['report-executive'],
    queryFn: reportService.executive,
  });

  if (isLoading) return <PageSpinner />;
  if (error || !data) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm p-4">
      <AlertCircle className="w-4 h-4" /> Không thể tải dữ liệu báo cáo
    </div>
  );

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><BarChart3 className="w-4 h-4 text-primary" />Tổng quan kinh doanh</CardTitle></CardHeader>
        <CardContent>
          <KpiRow label="Tổng doanh thu" value={formatCurrency(data.total_revenue)} />
          <KpiRow label="Số khách hàng" value={String(data.total_clients)} />
          <KpiRow label="Hợp đồng đang thực hiện" value={String(data.active_engagements)} />
          <KpiRow label="Công nợ phải thu" value={formatCurrency(data.outstanding_ar)} />
        </CardContent>
      </Card>

      {data.commission_kpis && (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><TrendingUp className="w-4 h-4 text-secondary" />Hoa hồng</CardTitle></CardHeader>
          <CardContent>
            <KpiRow label="Đã tích lũy" value={formatCurrency(data.commission_kpis.accrued)} />
            <KpiRow label="Đã chi" value={formatCurrency(data.commission_kpis.paid)} />
            <KpiRow label="Chờ duyệt" value={formatCurrency(data.commission_kpis.pending)} />
            <KpiRow label="Đang giữ" value={formatCurrency(data.commission_kpis.on_hold)} />
            <KpiRow
              label="% Doanh thu"
              value={`${(data.commission_kpis.commission_pct_of_revenue ?? 0).toFixed(1)}%`}
              sub="Commission / Revenue"
            />
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function ManagerReport() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['report-manager'],
    queryFn: reportService.manager,
  });

  if (isLoading) return <PageSpinner />;
  if (error || !data) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm p-4">
      <AlertCircle className="w-4 h-4" /> Không thể tải dữ liệu
    </div>
  );

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Card>
        <CardHeader><CardTitle className="flex items-center gap-2"><Users className="w-4 h-4 text-primary" />Dashboard Manager</CardTitle></CardHeader>
        <CardContent>
          <KpiRow label="Hợp đồng đang thực hiện" value={String(data.team_active_engagements)} />
          <KpiRow label="Chấm công chờ duyệt" value={String(data.team_pending_timesheets)} />
          <KpiRow label="Tỷ lệ sử dụng nhân lực" value={`${(data.team_utilization_rate * 100).toFixed(1)}%`} />
        </CardContent>
      </Card>
      {data.top_engagements?.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Briefcase className="w-4 h-4 text-primary" />Top hợp đồng</CardTitle></CardHeader>
          <CardContent>
            {data.top_engagements.map(eng => (
              <div key={eng.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{eng.title}</p>
                  <p className="text-xs text-text-secondary">{formatCurrency(eng.budget)}</p>
                </div>
                <Badge variant={eng.progress_pct >= 100 ? 'success' : 'default'} className="ml-2 flex-shrink-0">
                  {eng.progress_pct.toFixed(0)}%
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default function ReportsPage() {
  const { user } = useAuthStore();
  const isManager = user?.roles?.some(r => ['FIRM_PARTNER', 'AUDIT_MANAGER', 'SUPER_ADMIN'].includes(r));
  const [activeTab, setActiveTab] = React.useState<'overview' | 'commission' | 'billing'>('overview');

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">Báo cáo & Thống kê</h2>
        <p className="text-sm text-text-secondary">Tổng hợp dữ liệu từ materialized views — cập nhật mỗi đêm</p>
      </div>

      <div className="flex gap-2 border-b border-border">
        {[
          { key: 'overview', label: 'Tổng quan' },
          { key: 'commission', label: 'Hoa hồng' },
          { key: 'billing', label: 'Billing' },
        ].map(tab => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key as typeof activeTab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === tab.key ? 'border-action text-action' : 'border-transparent text-text-secondary hover:text-text-primary'}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'overview' && (isManager ? <ExecutiveReport /> : <ManagerReport />)}
      {activeTab === 'commission' && <CommissionReport />}
      {activeTab === 'billing' && <BillingReport />}
    </div>
  );
}

function CommissionReport() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['report-commission-payout'],
    queryFn: () => reportService.commissionPayout({ months: 6 }),
  });

  if (isLoading) return <PageSpinner />;
  if (error) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm p-4">
      <AlertCircle className="w-4 h-4" /> Không thể tải báo cáo hoa hồng
    </div>
  );

  return (
    <Card>
      <CardHeader><CardTitle className="flex items-center gap-2"><DollarSign className="w-4 h-4 text-secondary" />Báo cáo chi hoa hồng (6 tháng gần nhất)</CardTitle></CardHeader>
      <CardContent>
        {Array.isArray(data) && data.length > 0 ? (
          <div className="flex flex-col gap-2">
            {data.map((row: { month: string; total_approved: number; total_paid: number; record_count: number }) => (
              <div key={row.month} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                <span className="text-sm text-text-secondary font-mono">{row.month}</span>
                <div className="flex gap-6 text-right">
                  <div>
                    <p className="text-xs text-text-secondary">Đã duyệt</p>
                    <p className="text-sm font-medium">{formatCurrency(row.total_approved)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">Đã chi</p>
                    <p className="text-sm font-medium text-success">{formatCurrency(row.total_paid)}</p>
                  </div>
                  <div>
                    <p className="text-xs text-text-secondary">Giao dịch</p>
                    <p className="text-sm font-medium">{row.record_count}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-text-secondary text-center py-6">Không có dữ liệu</p>
        )}
      </CardContent>
    </Card>
  );
}

function BillingReport() {
  const today = new Date();
  const start = `${today.getFullYear()}-01-01`;
  const end = today.toISOString().split('T')[0];

  const { data, isLoading, error } = useQuery({
    queryKey: ['report-billing-period', start, end],
    queryFn: () => reportService.periodSummary(start, end),
  });

  if (isLoading) return <PageSpinner />;
  if (error) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm p-4">
      <AlertCircle className="w-4 h-4" /> Không thể tải báo cáo billing
    </div>
  );

  return (
    <Card>
      <CardHeader><CardTitle className="flex items-center gap-2"><BarChart3 className="w-4 h-4 text-primary" />Báo cáo billing năm {today.getFullYear()}</CardTitle></CardHeader>
      <CardContent>
        {data ? (
          <div className="flex flex-col gap-1">
            {Object.entries(data as Record<string, unknown>).map(([k, v]) => (
              <KpiRow key={k} label={k} value={typeof v === 'number' ? formatCurrency(v) : String(v)} />
            ))}
          </div>
        ) : (
          <p className="text-sm text-text-secondary text-center py-6">Không có dữ liệu</p>
        )}
      </CardContent>
    </Card>
  );
}
