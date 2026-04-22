'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/auth';
import { MODULE_ROLES, ROLE_GROUPS, hasAnyRole } from '@/lib/roles';
import { getModuleContext, MODULE_LABELS, type ModuleContext } from '@/lib/navigation';
import {
  LayoutDashboard, Users, Building2, Briefcase, Clock, FileText,
  DollarSign, FolderOpen, TrendingUp, BarChart3, Settings, LogOut,
  ShieldCheck, UserCog, GitBranch, ScrollText, Landmark,
  Layers, Grid3x3, Share2, UserCircle,
  UserPlus, ClipboardList, BookOpen, Award, Home,
} from 'lucide-react';

interface NavItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  roles?: string[];
}

interface ModuleNav {
  items: NavItem[];
}

const MODULE_NAVS: Record<Exclude<ModuleContext, null>, ModuleNav> = {
  hrm: {
    items: [
      { label: 'Dashboard',          href: '/hrm/dashboard',                      icon: LayoutDashboard },
      { label: 'Nhân viên',          href: '/admin/hrm/employees',                icon: Users,        roles: MODULE_ROLES.hrmEmployees },
      { label: 'Chi nhánh',          href: '/admin/hrm/organization/branches',    icon: Landmark,     roles: MODULE_ROLES.hrmOrgWrite },
      { label: 'Phòng ban',          href: '/admin/hrm/organization/departments', icon: Layers,       roles: MODULE_ROLES.hrmOrgWrite },
      { label: 'Ma trận',            href: '/admin/hrm/organization/matrix',      icon: Grid3x3,      roles: MODULE_ROLES.hrmOrgWrite },
      { label: 'Sơ đồ tổ chức',     href: '/admin/hrm/organization/org-chart',   icon: Share2,       roles: MODULE_ROLES.hrmOrg },
      { label: 'Danh mục khóa học',  href: '/admin/hrm/training-courses',         icon: BookOpen,     roles: MODULE_ROLES.hrmTrainingCourseRead },
      { label: 'Yêu cầu CPE',       href: '/admin/hrm/cpe-requirements',         icon: Award,        roles: MODULE_ROLES.hrmCPERead },
      { label: 'Cấp quyền',         href: '/admin/hrm/provisioning',             icon: UserPlus,     roles: MODULE_ROLES.hrmProvisioningRead },
      { label: 'Offboarding',        href: '/admin/hrm/offboarding',              icon: ClipboardList, roles: MODULE_ROLES.hrmOffboardingRead },
      { label: 'Hồ sơ của tôi',     href: '/my-profile',                         icon: UserCircle,   roles: ROLE_GROUPS.hr },
    ],
  },
  audit: {
    items: [
      { label: 'Tổng quan',         href: '/dashboard',       icon: LayoutDashboard, roles: [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit] },
      { label: 'Hợp đồng',          href: '/engagements',     icon: Briefcase,       roles: MODULE_ROLES.engagements },
      { label: 'Hồ sơ kiểm toán',   href: '/working-papers',  icon: FolderOpen,      roles: MODULE_ROLES.workingPapers },
      { label: 'Chấm công',         href: '/timesheets',      icon: Clock,           roles: MODULE_ROLES.timesheets },
      { label: 'Hoa hồng',          href: '/commissions',     icon: TrendingUp,      roles: MODULE_ROLES.commissions },
      { label: 'Hoa hồng của tôi',  href: '/commissions/my',  icon: TrendingUp,      roles: [...ROLE_GROUPS.audit] },
      { label: 'Hồ sơ của tôi',    href: '/my-profile',      icon: UserCircle },
    ],
  },
  finance: {
    items: [
      { label: 'Hóa đơn',    href: '/billing/invoices',  icon: FileText   },
      { label: 'Thanh toán', href: '/billing/payments',  icon: DollarSign },
    ],
  },
  crm: {
    items: [
      { label: 'Khách hàng', href: '/clients', icon: Building2 },
    ],
  },
  reports: {
    items: [
      { label: 'Báo cáo', href: '/reports', icon: BarChart3 },
    ],
  },
  system: {
    items: [
      { label: 'Dashboard',           href: '/admin/dashboard', icon: LayoutDashboard },
      { label: 'Người dùng & Vai trò', href: '/users',          icon: UserCog   },
      { label: 'Chi nhánh & Phòng ban', href: '/branches',      icon: GitBranch },
      { label: 'Nhật ký hệ thống',    href: '/audit-logs',      icon: ScrollText },
      { label: 'Cài đặt',             href: '/settings',        icon: Settings  },
    ],
  },
  client: {
    items: [
      { label: 'Cổng thông tin',    href: '/client/portal',    icon: LayoutDashboard },
      { label: 'Hợp đồng dịch vụ', href: '/engagements',      icon: Briefcase       },
      { label: 'Hóa đơn',          href: '/billing/invoices',  icon: FileText        },
      { label: 'Hồ sơ kiểm toán',  href: '/working-papers',   icon: FolderOpen      },
    ],
  },
};

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuthStore();
  const userRoles: string[] = user?.roles ?? [];

  const context = getModuleContext(pathname, userRoles);

  if (context === null) return null;

  const moduleNav = MODULE_NAVS[context];
  const visibleItems = moduleNav.items.filter(
    item => !item.roles || hasAnyRole(userRoles, item.roles)
  );

  return (
    <aside className="flex flex-col w-60 min-h-screen bg-primary text-white flex-shrink-0">
      {/* Logo */}
      <div className="flex items-center gap-3 px-5 py-5 border-b border-white/10">
        <div className="w-8 h-8 rounded bg-secondary/30 flex items-center justify-center flex-shrink-0">
          <ShieldCheck className="w-5 h-5 text-secondary" />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-bold leading-none truncate">ERP Audit</p>
          <p className="text-xs text-white/50 mt-0.5 truncate">v1.0</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 overflow-y-auto flex flex-col gap-1">
        {/* Home button */}
        <Link
          href="/"
          className="flex items-center gap-2 px-3 py-2 rounded text-sm text-white/60 hover:bg-white/10 hover:text-white transition-colors"
        >
          <Home className="w-4 h-4 flex-shrink-0" />
          Trang chủ
        </Link>

        {/* Divider + module label */}
        <div className="flex items-center gap-2 px-3 py-1 mt-1">
          <div className="h-px flex-1 bg-white/15" />
          <span className="text-xs font-semibold uppercase tracking-wider text-white/40">
            {MODULE_LABELS[context]}
          </span>
          <div className="h-px flex-1 bg-white/15" />
        </div>

        {/* Module nav items */}
        <ul className="flex flex-col gap-0.5 mt-1">
          {visibleItems.map(item => {
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

      {/* User footer */}
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
