'use client';
import { useAuthStore } from '@/stores/auth';
import {
  Users, Briefcase, DollarSign, Building2, BarChart3,
  Settings, Globe, Crown, ShieldCheck,
} from 'lucide-react';
import Link from 'next/link';

const MODULE_CARDS = [
  {
    label:   'HRM',
    href:    '/hrm/dashboard',
    icon:    Users,
    desc:    'Nhân sự, lương, chứng chỉ, đào tạo',
    color:   'bg-blue-50 text-blue-600',
  },
  {
    label:   'Kiểm toán',
    href:    '/engagements',
    icon:    Briefcase,
    desc:    'Hợp đồng, hồ sơ kiểm toán, chấm công',
    color:   'bg-indigo-50 text-indigo-600',
  },
  {
    label:   'Tài chính',
    href:    '/billing/invoices',
    icon:    DollarSign,
    desc:    'Hóa đơn, thanh toán, công nợ',
    color:   'bg-emerald-50 text-emerald-600',
  },
  {
    label:   'CRM',
    href:    '/clients',
    icon:    Building2,
    desc:    'Danh sách và hồ sơ khách hàng',
    color:   'bg-orange-50 text-orange-600',
  },
  {
    label:   'Báo cáo',
    href:    '/reports',
    icon:    BarChart3,
    desc:    'Báo cáo tổng hợp toàn công ty',
    color:   'bg-purple-50 text-purple-600',
  },
  {
    label:   'Hệ thống',
    href:    '/users',
    icon:    Settings,
    desc:    'Người dùng, vai trò, cài đặt hệ thống',
    color:   'bg-slate-50 text-slate-600',
  },
  {
    label:   'Dịch vụ khách hàng',
    href:    '/client/portal',
    icon:    Globe,
    desc:    'Cổng thông tin dành cho khách hàng',
    color:   'bg-teal-50 text-teal-600',
  },
  {
    label:   'Executive',
    href:    '/executive/dashboard',
    icon:    Crown,
    desc:    'Tổng quan điều hành dành cho lãnh đạo',
    color:   'bg-yellow-50 text-yellow-600',
  },
];

export default function AdminDashboardPage() {
  const { user } = useAuthStore();

  return (
    <div className="max-w-5xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Quản trị hệ thống</h1>
        <p className="text-sm text-text-secondary mt-1">
          Xin chào, {user?.full_name}. Chọn module để điều hướng nhanh.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {MODULE_CARDS.map(({ label, href, icon: Icon, desc, color }) => (
          <Link
            key={href}
            href={href}
            className="flex flex-col gap-3 p-5 rounded-card bg-white border border-border hover:border-primary/40 hover:shadow-sm transition-all group"
          >
            <div className={`w-10 h-10 rounded flex items-center justify-center flex-shrink-0 ${color}`}>
              <Icon className="w-5 h-5" />
            </div>
            <div>
              <p className="text-sm font-semibold text-text-primary group-hover:text-primary transition-colors">
                {label}
              </p>
              <p className="text-xs text-text-secondary mt-0.5 leading-relaxed">{desc}</p>
            </div>
          </Link>
        ))}
      </div>

      <div className="rounded-card bg-white border border-border p-4">
        <div className="flex items-center gap-2 mb-2">
          <ShieldCheck className="w-4 h-4 text-primary" />
          <h2 className="text-sm font-semibold text-text-primary">Lưu ý bảo mật</h2>
        </div>
        <p className="text-xs text-text-secondary">
          Với quyền SUPER_ADMIN, bạn có thể truy cập tất cả module. Mọi thao tác đều được ghi vào audit log.
        </p>
      </div>
    </div>
  );
}
