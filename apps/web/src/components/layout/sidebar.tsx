'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/auth';
import {
  LayoutDashboard, Users, Building2, Briefcase, Clock, FileText,
  DollarSign, FolderOpen, TrendingUp, BarChart3, Settings, LogOut, ShieldCheck, UserCog, GitBranch, ScrollText,
} from 'lucide-react';

interface NavItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  roles?: string[];
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { label: 'Khách hàng', href: '/clients', icon: Building2 },
  { label: 'Nhân viên', href: '/employees', icon: Users, roles: ['FIRM_PARTNER', 'AUDIT_MANAGER', 'SUPER_ADMIN'] },
  { label: 'Người dùng & Vai trò', href: '/users', icon: UserCog, roles: ['SUPER_ADMIN', 'FIRM_PARTNER'] },
  { label: 'Chi nhánh & Phòng ban', href: '/branches', icon: GitBranch, roles: ['SUPER_ADMIN', 'FIRM_PARTNER'] },
  { label: 'Hợp đồng', href: '/engagements', icon: Briefcase },
  { label: 'Chấm công', href: '/timesheets', icon: Clock },
  { label: 'Hóa đơn', href: '/billing/invoices', icon: FileText },
  { label: 'Thanh toán', href: '/billing/payments', icon: DollarSign },
  { label: 'Hồ sơ kiểm toán', href: '/working-papers', icon: FolderOpen },
  { label: 'Hoa hồng', href: '/commissions', icon: TrendingUp },
  { label: 'Báo cáo', href: '/reports', icon: BarChart3, roles: ['FIRM_PARTNER', 'AUDIT_MANAGER', 'SUPER_ADMIN'] },
  { label: 'Nhật ký hệ thống', href: '/audit-logs', icon: ScrollText, roles: ['SUPER_ADMIN'] },
  { label: 'Cài đặt', href: '/settings', icon: Settings, roles: ['SUPER_ADMIN'] },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuthStore();

  const visibleItems = NAV_ITEMS.filter(item => {
    if (!item.roles) return true;
    return item.roles.some(r => user?.roles?.includes(r));
  });

  return (
    <aside className="flex flex-col w-60 min-h-screen bg-primary text-white">
      <div className="flex items-center gap-3 px-5 py-5 border-b border-white/10">
        <div className="w-8 h-8 rounded bg-secondary/30 flex items-center justify-center flex-shrink-0">
          <ShieldCheck className="w-5 h-5 text-secondary" />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-bold leading-none truncate">ERP Audit</p>
          <p className="text-xs text-white/50 mt-0.5 truncate">v1.0</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 overflow-y-auto">
        <ul className="flex flex-col gap-0.5">
          {visibleItems.map((item) => {
            const active = pathname === item.href || pathname.startsWith(item.href + '/');
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={cn(
                    'flex items-center gap-3 px-3 py-2 rounded text-sm transition-colors',
                    active
                      ? 'bg-white/15 text-white font-medium'
                      : 'text-white/70 hover:bg-white/10 hover:text-white'
                  )}
                >
                  <item.icon className="w-4 h-4 flex-shrink-0" />
                  {item.label}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="px-3 py-3 border-t border-white/10">
        <div className="flex items-center gap-3 px-3 py-2 mb-1">
          <div className="w-7 h-7 rounded-full bg-secondary/30 flex items-center justify-center flex-shrink-0">
            <span className="text-xs font-bold text-secondary">
              {user?.full_name?.[0]?.toUpperCase() ?? 'U'}
            </span>
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-xs font-medium text-white truncate">{user?.full_name ?? 'User'}</p>
            <p className="text-xs text-white/50 truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={logout}
          className="flex w-full items-center gap-3 px-3 py-2 rounded text-sm text-white/70 hover:bg-white/10 hover:text-white transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Đăng xuất
        </button>
      </div>
    </aside>
  );
}
