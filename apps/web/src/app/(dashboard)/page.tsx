'use client';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/auth';
import { hasAnyRole, ROLE_GROUPS, MODULE_ROLES } from '@/lib/roles';
import { reportService } from '@/services/reports';
import { formatCurrency } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Users, Briefcase, DollarSign, Building2, BarChart3,
  Settings, LayoutDashboard, AlertCircle,
} from 'lucide-react';

// ── Executive KPI ────────────────────────────────────────────────────────────

function StatCard({ title, value, icon: Icon }: {
  title: string;
  value: string;
  icon: React.ComponentType<{ className?: string }>;
}) {
  return (
    <Card>
      <CardContent className="flex items-center gap-4 pt-5">
        <div className="w-10 h-10 rounded-card bg-primary/10 flex items-center justify-center flex-shrink-0">
          <Icon className="w-5 h-5 text-primary" />
        </div>
        <div className="min-w-0">
          <p className="text-xs text-text-secondary">{title}</p>
          <p className="text-xl font-bold text-text-primary mt-0.5">{value}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function ExecutiveKPI() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard', 'executive'],
    queryFn: reportService.executive,
  });

  if (isLoading) return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="h-24 rounded-card bg-border" />
      ))}
    </div>
  );

  if (error || !data) return (
    <div className="flex items-center gap-2 text-text-secondary text-sm">
      <AlertCircle className="w-4 h-4" /> Không thể tải dữ liệu KPI
    </div>
  );

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard title="Doanh thu"             value={formatCurrency(data.total_revenue)}     icon={DollarSign} />
        <StatCard title="Khách hàng"            value={String(data.total_clients)}             icon={Users}      />
        <StatCard title="Hợp đồng đang thực hiện" value={String(data.active_engagements)}     icon={Briefcase}  />
        <StatCard title="Công nợ phải thu"      value={formatCurrency(data.outstanding_ar)}   icon={BarChart3}  />
      </div>

      {data.commission_kpis && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Hoa hồng tổng quan</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
              {[
                { label: 'Đã tích lũy', value: data.commission_kpis.accrued },
                { label: 'Đã chi',      value: data.commission_kpis.paid    },
                { label: 'Chờ duyệt',   value: data.commission_kpis.pending },
                { label: 'Đang giữ',    value: data.commission_kpis.on_hold },
                { label: '% Doanh thu', pct: data.commission_kpis.commission_pct_of_revenue },
              ].map((item) => (
                <div key={item.label} className="flex flex-col gap-1">
                  <p className="text-xs text-text-secondary">{item.label}</p>
                  <p className="text-sm font-semibold text-text-primary">
                    {'pct' in item
                      ? `${item.pct?.toFixed(1)}%`
                      : formatCurrency(item.value ?? 0)}
                  </p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

// ── Module card definitions ───────────────────────────────────────────────────

const MODULE_CARDS = [
  {
    key: 'hrm',
    label: 'Nhân sự',
    desc:  'Quản lý nhân viên, tổ chức, đào tạo & cấp quyền',
    href:  '/hrm/dashboard',
    icon:  Users,
    roles: [...ROLE_GROUPS.hr, ...ROLE_GROUPS.sysAdmin, ...ROLE_GROUPS.executive],
    color: 'bg-blue-500/10 text-blue-600',
  },
  {
    key: 'audit',
    label: 'Kiểm toán',
    desc:  'Hợp đồng, hồ sơ kiểm toán, chấm công & hoa hồng',
    href:  '/engagements',
    icon:  Briefcase,
    roles: [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit, ...ROLE_GROUPS.sysAdmin, ...ROLE_GROUPS.executive],
    color: 'bg-violet-500/10 text-violet-600',
  },
  {
    key: 'finance',
    label: 'Tài chính',
    desc:  'Hóa đơn, thanh toán và công nợ phải thu',
    href:  '/billing/invoices',
    icon:  DollarSign,
    roles: MODULE_ROLES.billing,
    color: 'bg-emerald-500/10 text-emerald-600',
  },
  {
    key: 'crm',
    label: 'CRM',
    desc:  'Quản lý khách hàng và hợp đồng dịch vụ',
    href:  '/clients',
    icon:  Building2,
    roles: [...ROLE_GROUPS.sysAdmin, ...ROLE_GROUPS.executive, ...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit],
    color: 'bg-amber-500/10 text-amber-600',
  },
  {
    key: 'reports',
    label: 'Báo cáo',
    desc:  'Thống kê và báo cáo tổng hợp toàn công ty',
    href:  '/reports',
    icon:  BarChart3,
    roles: MODULE_ROLES.reports,
    color: 'bg-rose-500/10 text-rose-600',
  },
  {
    key: 'system',
    label: 'Hệ thống',
    desc:  'Quản trị người dùng, vai trò và cài đặt hệ thống',
    href:  '/admin/dashboard',
    icon:  Settings,
    roles: ROLE_GROUPS.sysAdmin,
    color: 'bg-slate-500/10 text-slate-600',
  },
  {
    key: 'client',
    label: 'Dịch vụ',
    desc:  'Cổng thông tin khách hàng, hồ sơ & hóa đơn',
    href:  '/client/portal',
    icon:  LayoutDashboard,
    roles: ROLE_GROUPS.client,
    color: 'bg-cyan-500/10 text-cyan-600',
  },
] as const;

// ── Home page ─────────────────────────────────────────────────────────────────

export default function HomePage() {
  const { user } = useAuthStore();
  const userRoles: string[] = user?.roles ?? [];
  const isExecutive = hasAnyRole(userRoles, ROLE_GROUPS.executive);

  const isSuperAdmin = hasAnyRole(userRoles, ROLE_GROUPS.sysAdmin);
  const visibleCards = MODULE_CARDS.filter(card => {
    // client card chỉ dành cho client roles
    if (card.key === 'client') return hasAnyRole(userRoles, ROLE_GROUPS.client);
    // SUPER_ADMIN thấy tất cả module còn lại
    return isSuperAdmin || hasAnyRole(userRoles, [...card.roles]);
  });

  return (
    <div className="max-w-5xl mx-auto space-y-8">
      {/* Welcome */}
      <div>
        <h1 className="text-2xl font-bold text-text-primary">
          Xin chào, {user?.full_name ?? 'bạn'}!
        </h1>
        <p className="text-sm text-text-secondary mt-1">
          Chọn module bạn muốn làm việc hôm nay
        </p>
      </div>

      {/* KPI section — executives only */}
      {isExecutive && (
        <section className="space-y-2">
          <h2 className="text-sm font-semibold text-text-secondary uppercase tracking-wider">
            Tổng quan điều hành
          </h2>
          <ExecutiveKPI />
        </section>
      )}

      {/* Module cards */}
      <section className="space-y-3">
        {isExecutive && (
          <h2 className="text-sm font-semibold text-text-secondary uppercase tracking-wider">
            Module
          </h2>
        )}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {visibleCards.map(({ key, label, desc, href, icon: Icon, color }) => (
            <Link
              key={key}
              href={href}
              className="flex items-start gap-4 p-5 rounded-card bg-white border border-border hover:border-primary/30 hover:shadow-sm transition-all group"
            >
              <div className={`w-11 h-11 rounded-lg flex items-center justify-center flex-shrink-0 ${color}`}>
                <Icon className="w-6 h-6" />
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-text-primary group-hover:text-primary transition-colors">
                  {label}
                </p>
                <p className="text-xs text-text-secondary mt-0.5 leading-relaxed">{desc}</p>
              </div>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}
