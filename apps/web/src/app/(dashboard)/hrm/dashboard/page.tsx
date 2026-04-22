'use client';
import { useAuthStore } from '@/stores/auth';
import { Users, Landmark, Share2, UserCircle } from 'lucide-react';
import Link from 'next/link';
import { ROLES } from '@/lib/roles';

export default function HrmDashboardPage() {
  const { user } = useAuthStore();
  const userRoles: string[] = user?.roles ?? [];

  const isHrManager = userRoles.includes(ROLES.HR_MANAGER);
  const isHoB = userRoles.includes(ROLES.HEAD_OF_BRANCH);

  const links = [
    ...(isHrManager || userRoles.includes(ROLES.HR_STAFF) ? [{
      label: 'Nhân viên', href: '/admin/hrm/employees', icon: Users, desc: 'Danh sách và hồ sơ nhân viên',
    }] : []),
    ...(isHoB ? [{
      label: 'Nhân viên chi nhánh', href: '/admin/hrm/employees', icon: Users, desc: 'Nhân viên trong chi nhánh của bạn',
    }] : []),
    ...(isHrManager ? [
      { label: 'Cơ cấu tổ chức', href: '/admin/hrm/organization/org-chart', icon: Share2, desc: 'Sơ đồ tổ chức công ty' },
      { label: 'Chi nhánh', href: '/admin/hrm/organization/branches', icon: Landmark, desc: 'Quản lý chi nhánh' },
    ] : []),
    { label: 'Hồ sơ của tôi', href: '/my-profile', icon: UserCircle, desc: 'Xem và cập nhật thông tin cá nhân' },
  ];

  return (
    <div className="max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Nhân sự (HRM)</h1>
        <p className="text-sm text-text-secondary mt-1">Xin chào, {user?.full_name}.</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {links.map(({ label, href, icon: Icon, desc }) => (
          <Link
            key={`${href}-${label}`}
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
