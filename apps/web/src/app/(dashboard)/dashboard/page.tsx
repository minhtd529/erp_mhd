'use client';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { PageSpinner } from '@/components/ui/spinner';
import { Badge } from '@/components/ui/badge';
import { reportService } from '@/services/reports';
import { useAuthStore } from '@/stores/auth';
import { formatCurrency } from '@/lib/utils';
import { DollarSign, Users, Briefcase, TrendingUp, AlertCircle } from 'lucide-react';

function StatCard({ title, value, icon: Icon, sub }: { title: string; value: string; icon: React.ComponentType<{className?: string}>; sub?: string }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-4 pt-5">
        <div className="w-10 h-10 rounded-card bg-primary/10 flex items-center justify-center flex-shrink-0">
          <Icon className="w-5 h-5 text-primary" />
        </div>
        <div className="min-w-0">
          <p className="text-xs text-text-secondary">{title}</p>
          <p className="text-xl font-bold text-text-primary mt-0.5">{value}</p>
          {sub && <p className="text-xs text-text-secondary">{sub}</p>}
        </div>
      </CardContent>
    </Card>
  );
}

function ExecutiveDashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard', 'executive'],
    queryFn: reportService.executive,
  });

  if (isLoading) return <PageSpinner />;
  if (error || !data) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm">
      <AlertCircle className="w-4 h-4" /> Không thể tải dữ liệu dashboard
    </div>
  );

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard title="Doanh thu" value={formatCurrency(data.total_revenue)} icon={DollarSign} />
        <StatCard title="Khách hàng" value={String(data.total_clients)} icon={Users} />
        <StatCard title="Hợp đồng đang thực hiện" value={String(data.active_engagements)} icon={Briefcase} />
        <StatCard title="Công nợ phải thu" value={formatCurrency(data.outstanding_ar)} icon={TrendingUp} />
      </div>

      {data.commission_kpis && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle>Hoa hồng tổng quan</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
              {[
                { label: 'Đã tích lũy', value: data.commission_kpis.accrued, variant: 'default' as const },
                { label: 'Đã chi', value: data.commission_kpis.paid, variant: 'success' as const },
                { label: 'Chờ duyệt', value: data.commission_kpis.pending, variant: 'warning' as const },
                { label: 'Đang giữ', value: data.commission_kpis.on_hold, variant: 'secondary' as const },
                { label: '% Doanh thu', value: null, pct: data.commission_kpis.commission_pct_of_revenue, variant: 'outline' as const },
              ].map((item) => (
                <div key={item.label} className="flex flex-col gap-1">
                  <p className="text-xs text-text-secondary">{item.label}</p>
                  <p className="text-sm font-semibold text-text-primary">
                    {item.pct !== undefined ? `${item.pct.toFixed(1)}%` : formatCurrency(item.value ?? 0)}
                  </p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </>
  );
}

function PersonalDashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard', 'personal'],
    queryFn: reportService.personal,
  });

  if (isLoading) return <PageSpinner />;
  if (error || !data) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm">
      <AlertCircle className="w-4 h-4" /> Không thể tải dữ liệu cá nhân
    </div>
  );

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard title="Hợp đồng đang tham gia" value={String(data.active_engagements)} icon={Briefcase} />
        <StatCard title="Chấm công chờ duyệt" value={String(data.pending_timesheets)} icon={AlertCircle} />
        <StatCard title="Giờ làm tháng này" value={`${data.total_hours_this_month}h`} icon={TrendingUp} />
      </div>
      {data.is_salesperson && (
        <Card className="mt-4">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              Hoa hồng của tôi <Badge variant="success">Salesperson</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              {[
                { label: 'Cả năm (YTD)', value: data.commission_ytd },
                { label: 'Tháng này', value: data.commission_month },
                { label: 'Chờ duyệt', value: data.commission_pending },
                { label: 'Đang giữ', value: data.commission_on_hold },
              ].map((item) => (
                <div key={item.label}>
                  <p className="text-xs text-text-secondary">{item.label}</p>
                  <p className="text-sm font-semibold text-text-primary">{formatCurrency(item.value ?? 0)}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </>
  );
}

export default function DashboardPage() {
  const { user } = useAuthStore();
  const isManager = user?.roles?.some(r => ['FIRM_PARTNER', 'AUDIT_MANAGER', 'SUPER_ADMIN'].includes(r));

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-semibold text-text-primary">
          Xin chào, {user?.full_name ?? 'bạn'} 👋
        </h2>
        <p className="text-sm text-text-secondary">Đây là tổng quan hệ thống hôm nay</p>
      </div>
      {isManager ? <ExecutiveDashboard /> : <PersonalDashboard />}
    </div>
  );
}
