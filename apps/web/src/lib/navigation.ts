import { ROLE_GROUPS, hasAnyRole } from './roles';

export type ModuleContext =
  | 'hrm' | 'audit' | 'finance' | 'crm' | 'reports' | 'system' | 'client'
  | null;

export const MODULE_LABELS: Record<Exclude<ModuleContext, null>, string> = {
  hrm:     'HRM',
  audit:   'Kiểm toán',
  finance: 'Tài chính',
  crm:     'CRM',
  reports: 'Báo cáo',
  system:  'Hệ thống',
  client:  'Dịch vụ',
};

export const MODULE_HOME: Record<Exclude<ModuleContext, null>, string> = {
  hrm:     '/hrm/dashboard',
  audit:   '/engagements',
  finance: '/billing/invoices',
  crm:     '/clients',
  reports: '/reports',
  system:  '/admin/dashboard',
  client:  '/client/portal',
};

const AUDIT_PREFIXES = ['/engagements', '/working-papers', '/timesheets', '/commissions'];
const SYSTEM_PREFIXES = ['/users', '/branches', '/audit-logs', '/settings'];

export function getModuleContext(pathname: string, userRoles: string[]): ModuleContext {
  if (pathname === '/') return null;
  if (pathname.startsWith('/hrm') || pathname.startsWith('/admin/hrm')) return 'hrm';
  if (pathname.startsWith('/admin')) return 'system';
  if (AUDIT_PREFIXES.some(p => pathname === p || pathname.startsWith(p + '/'))) return 'audit';
  if (pathname.startsWith('/billing')) return 'finance';
  if (pathname.startsWith('/clients')) return 'crm';
  if (pathname.startsWith('/client/')) {
    return hasAnyRole(userRoles, ROLE_GROUPS.client) ? 'client' : 'crm';
  }
  if (pathname.startsWith('/reports')) return 'reports';
  if (SYSTEM_PREFIXES.some(p => pathname === p || pathname.startsWith(p + '/'))) return 'system';
  if (pathname.startsWith('/my-profile')) {
    return hasAnyRole(userRoles, ROLE_GROUPS.hr) ? 'hrm' : 'audit';
  }
  // /dashboard and other ambiguous paths — infer from role
  if (hasAnyRole(userRoles, ROLE_GROUPS.client)) return 'client';
  if (hasAnyRole(userRoles, ROLE_GROUPS.hr)) return 'hrm';
  if (hasAnyRole(userRoles, [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit])) return 'audit';
  return 'system';
}

// Ordered longest-first so more specific prefixes match before shorter ones
const PAGE_LABELS: { prefix: string; label: string }[] = [
  { prefix: '/hrm/dashboard',                       label: 'Dashboard' },
  { prefix: '/admin/hrm/employees',                 label: 'Nhân viên' },
  { prefix: '/admin/hrm/organization/branches',     label: 'Chi nhánh' },
  { prefix: '/admin/hrm/organization/departments',  label: 'Phòng ban' },
  { prefix: '/admin/hrm/organization/matrix',       label: 'Ma trận' },
  { prefix: '/admin/hrm/organization/org-chart',    label: 'Sơ đồ tổ chức' },
  { prefix: '/admin/hrm/organization',              label: 'Tổ chức' },
  { prefix: '/admin/hrm/training-courses',          label: 'Danh mục khóa học' },
  { prefix: '/admin/hrm/cpe-requirements',          label: 'Yêu cầu CPE' },
  { prefix: '/admin/hrm/provisioning',              label: 'Cấp quyền' },
  { prefix: '/admin/hrm/offboarding',               label: 'Offboarding' },
  { prefix: '/admin/dashboard',                     label: 'Dashboard' },
  { prefix: '/engagements',                         label: 'Hợp đồng' },
  { prefix: '/working-papers',                      label: 'Hồ sơ kiểm toán' },
  { prefix: '/timesheets',                          label: 'Chấm công' },
  { prefix: '/commissions/my',                      label: 'Hoa hồng của tôi' },
  { prefix: '/commissions',                         label: 'Hoa hồng' },
  { prefix: '/billing/invoices',                    label: 'Hóa đơn' },
  { prefix: '/billing/payments',                    label: 'Thanh toán' },
  { prefix: '/clients',                             label: 'Khách hàng' },
  { prefix: '/client/portal',                       label: 'Cổng thông tin' },
  { prefix: '/reports',                             label: 'Báo cáo' },
  { prefix: '/users',                               label: 'Người dùng & Vai trò' },
  { prefix: '/branches',                            label: 'Chi nhánh & Phòng ban' },
  { prefix: '/audit-logs',                          label: 'Nhật ký hệ thống' },
  { prefix: '/settings',                            label: 'Cài đặt' },
  { prefix: '/my-profile',                          label: 'Hồ sơ của tôi' },
  { prefix: '/dashboard',                           label: 'Tổng quan' },
];

export function getPageLabel(pathname: string): string {
  for (const { prefix, label } of PAGE_LABELS) {
    if (pathname === prefix || pathname.startsWith(prefix + '/')) return label;
  }
  return '';
}
