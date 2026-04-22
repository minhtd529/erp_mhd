'use client';
import { useAuthStore } from '@/stores/auth';
import { Users, Briefcase, DollarSign, BarChart3 } from 'lucide-react';
import Link from 'next/link';

const QUICK_LINKS = [
  { label: 'Nhân sự', href: '/admin/hrm/employees', icon: Users, desc: 'Xem danh sách nhân viên toàn công ty' },
  { label: 'Hợp đồng kiểm toán', href: '/engagements', icon: Briefcase, desc: 'Theo dõi các engagement đang thực hiện' },
  { label: 'Tài chính', href: '/billing/invoices', icon: DollarSign, desc: 'Hóa đơn và thanh toán' },
  { label: 'Báo cáo', href: '/reports', icon: BarChart3, desc: 'Báo cáo tổng hợp toàn công ty' },
];

export default function ExecutiveDashboardPage() {
  const { user } = useAuthStore();
  const role = user?.roles?.includes('CHAIRMAN') ? 'Chairman' : 'CEO';

  return (
    <div className="max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Tổng quan điều hành</h1>
        <p className="text-sm text-text-secondary mt-1">
          Xin chào, {user?.full_name}. Dashboard dành cho {role}.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {QUICK_LINKS.map(({ label, href, icon: Icon, desc }) => (
          <Link
            key={href}
            href={href}
            className="flex items-start gap-4 p-4 rounded-card bg-white border border-border hover:border-primary/40 hover:shadow-sm transition-all"
          >
            <div className="w-10 h-10 rounded bg-primary/10 flex items-center justify-center flex-shrink-0">
              <Icon className="w-5 h-5 text-primary" />
            </div>
            <div>
              <p className="text-sm font-semibold text-text-primary">{label}</p>
              <p className="text-xs text-text-secondary mt-0.5">{desc}</p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
