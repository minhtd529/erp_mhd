'use client';
import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/auth';
import { ROLE_GROUPS, MODULE_ROLES, hasAnyRole } from '@/lib/roles';
import {
  LayoutDashboard, Users, Building2, Briefcase, Clock, FileText,
  DollarSign, FolderOpen, TrendingUp, BarChart3, Settings, LogOut,
  ShieldCheck, UserCog, GitBranch, ScrollText, Network, Landmark,
  Layers, Grid3x3, Share2, UserCircle, ChevronDown, ChevronRight,
  UserPlus, ClipboardList,
} from 'lucide-react';

interface NavItem {
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  roles?: string[];
}

interface NavGroup {
  title: string;
  roles?: string[];         // Group only shows if user has any of these roles
  defaultOpen?: boolean;
  items: NavItem[];
}

const NAV_GROUPS: NavGroup[] = [
  // ── Quản trị hệ thống (SUPER_ADMIN only) ────────────────────────────
  {
    title: 'Quản trị hệ thống',
    roles: ROLE_GROUPS.sysAdmin,
    defaultOpen: true,
    items: [
      { label: 'Admin Dashboard', href: '/admin/dashboard', icon: LayoutDashboard },
      { label: 'Người dùng & Vai trò', href: '/users', icon: UserCog },
      { label: 'Chi nhánh & Phòng ban', href: '/branches', icon: GitBranch },
      { label: 'Nhật ký hệ thống', href: '/audit-logs', icon: ScrollText },
      { label: 'Cài đặt', href: '/settings', icon: Settings },
    ],
  },

  // ── Tổng quan (partner + audit + executive) ──────────────────────────
  {
    title: 'Tổng quan',
    roles: [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit],
    defaultOpen: true,
    items: [
      { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
    ],
  },

  // ── Tổng quan điều hành (CHAIRMAN, CEO) ─────────────────────────────
  {
    title: 'Tổng quan',
    roles: ROLE_GROUPS.executive,
    defaultOpen: true,
    items: [
      { label: 'Dashboard điều hành', href: '/executive/dashboard', icon: LayoutDashboard },
    ],
  },

  // ── Tổng quan nhân sự (HR roles) ─────────────────────────────────────
  {
    title: 'Tổng quan',
    roles: ROLE_GROUPS.hr,
    defaultOpen: true,
    items: [
      { label: 'HRM Dashboard', href: '/hrm/dashboard', icon: LayoutDashboard },
    ],
  },

  // ── CRM ──────────────────────────────────────────────────────────────
  {
    title: 'CRM',
    roles: [...ROLE_GROUPS.sysAdmin, ...ROLE_GROUPS.executive, ...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit],
    defaultOpen: true,
    items: [
      { label: 'Khách hàng', href: '/clients', icon: Building2 },
    ],
  },

  // ── Hợp đồng kiểm toán ──────────────────────────────────────────────
  {
    title: 'Hợp đồng kiểm toán',
    roles: MODULE_ROLES.engagements,
    defaultOpen: true,
    items: [
      { label: 'Hợp đồng', href: '/engagements', icon: Briefcase },
      { label: 'Hồ sơ kiểm toán', href: '/working-papers', icon: FolderOpen, roles: MODULE_ROLES.workingPapers },
    ],
  },

  // ── Timesheet & Hoa hồng (audit staff + partner) ────────────────────
  {
    title: 'Công việc cá nhân',
    roles: [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit],
    defaultOpen: true,
    items: [
      { label: 'Chấm công', href: '/timesheets', icon: Clock },
      { label: 'Hoa hồng', href: '/commissions', icon: TrendingUp, roles: MODULE_ROLES.commissions },
      { label: 'Hoa hồng của tôi', href: '/commissions/my', icon: TrendingUp, roles: [...ROLE_GROUPS.audit] },
      { label: 'Hồ sơ của tôi', href: '/my-profile', icon: UserCircle },
    ],
  },

  // ── Tài chính ────────────────────────────────────────────────────────
  {
    title: 'Tài chính',
    roles: MODULE_ROLES.billing,
    defaultOpen: false,
    items: [
      { label: 'Hóa đơn', href: '/billing/invoices', icon: FileText },
      { label: 'Thanh toán', href: '/billing/payments', icon: DollarSign },
    ],
  },

  // ── Nhân sự (HRM) — org section ─────────────────────────────────────
  {
    title: 'Tổ chức',
    roles: MODULE_ROLES.hrmOrgWrite,
    defaultOpen: false,
    items: [
      { label: 'Chi nhánh', href: '/admin/hrm/organization/branches', icon: Landmark },
      { label: 'Phòng ban', href: '/admin/hrm/organization/departments', icon: Layers },
      { label: 'Ma trận', href: '/admin/hrm/organization/matrix', icon: Grid3x3 },
      { label: 'Sơ đồ tổ chức', href: '/admin/hrm/organization/org-chart', icon: Share2 },
    ],
  },

  // ── Nhân sự (HRM) — employee section ────────────────────────────────
  {
    title: 'Nhân viên',
    roles: MODULE_ROLES.hrmEmployees,
    defaultOpen: true,
    items: [
      { label: 'Danh sách nhân viên', href: '/admin/hrm/employees', icon: Users },
    ],
  },

  // ── Nhân sự (HRM) — provisioning & offboarding ───────────────────────
  {
    title: 'Cấp quyền & Offboarding',
    roles: MODULE_ROLES.hrmProvisioningRead,
    defaultOpen: true,
    items: [
      {
        label: 'Cấp tài khoản',
        href: '/admin/hrm/provisioning',
        icon: UserPlus,
        roles: MODULE_ROLES.hrmProvisioningRead,
      },
      {
        label: 'Offboarding',
        href: '/admin/hrm/offboarding',
        icon: ClipboardList,
        roles: MODULE_ROLES.hrmOffboardingRead,
      },
    ],
  },

  // ── Hồ sơ cá nhân — HR roles ─────────────────────────────────────────
  {
    title: 'Cá nhân',
    roles: ROLE_GROUPS.hr,
    defaultOpen: true,
    items: [
      { label: 'Hồ sơ của tôi', href: '/my-profile', icon: UserCircle },
    ],
  },

  // ── Báo cáo ──────────────────────────────────────────────────────────
  {
    title: 'Báo cáo',
    roles: MODULE_ROLES.reports,
    defaultOpen: false,
    items: [
      { label: 'Báo cáo', href: '/reports', icon: BarChart3 },
    ],
  },

  // ── Client portal ────────────────────────────────────────────────────
  {
    title: 'Dịch vụ',
    roles: ROLE_GROUPS.client,
    defaultOpen: true,
    items: [
      { label: 'Cổng thông tin', href: '/client/portal', icon: LayoutDashboard },
      { label: 'Hợp đồng dịch vụ', href: '/engagements', icon: Briefcase },
      { label: 'Hóa đơn', href: '/billing/invoices', icon: FileText },
      { label: 'Hồ sơ kiểm toán', href: '/working-papers', icon: FolderOpen },
    ],
  },
];

function NavGroupSection({ group, userRoles, pathname }: {
  group: NavGroup;
  userRoles: string[];
  pathname: string;
}) {
  const [open, setOpen] = useState(group.defaultOpen ?? true);

  const visibleItems = group.items.filter(item =>
    !item.roles || hasAnyRole(userRoles, item.roles)
  );

  if (visibleItems.length === 0) return null;

  return (
    <div className="mb-1">
      <button
        onClick={() => setOpen(v => !v)}
        className="flex items-center justify-between w-full px-3 py-1.5 text-xs font-semibold uppercase tracking-wider text-white/40 hover:text-white/60 transition-colors"
      >
        <span>{group.title}</span>
        {open
          ? <ChevronDown className="w-3 h-3" />
          : <ChevronRight className="w-3 h-3" />
        }
      </button>

      {open && (
        <ul className="flex flex-col gap-0.5">
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
      )}
    </div>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout } = useAuthStore();
  const userRoles: string[] = user?.roles ?? [];

  const visibleGroups = NAV_GROUPS.filter(group =>
    !group.roles || hasAnyRole(userRoles, group.roles)
  );

  // Deduplicate groups that have the same title (e.g. multiple "Tổng quan" definitions)
  // — only the first matching one per title renders
  const seenTitles = new Set<string>();
  const deduped = visibleGroups.filter(group => {
    // Groups without roles collision risk don't need dedup
    if (!group.roles) return true;
    const key = group.title + '|' + group.items.map(i => i.href).join(',');
    if (seenTitles.has(group.title)) {
      // Allow if different items
      if (seenTitles.has(key)) return false;
    }
    seenTitles.add(group.title);
    seenTitles.add(key);
    return true;
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
        {deduped.map((group, idx) => (
          <NavGroupSection
            key={`${group.title}-${idx}`}
            group={group}
            userRoles={userRoles}
            pathname={pathname}
          />
        ))}
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
