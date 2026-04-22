'use client';
import { useAuthStore } from '@/stores/auth';
import { ShieldCheck, Users, ScrollText, Settings, Building2 } from 'lucide-react';
import Link from 'next/link';

const QUICK_LINKS = [
  { label: 'Người dùng & Vai trò', href: '/users', icon: Users, desc: 'Quản lý tài khoản và phân quyền' },
  { label: 'Chi nhánh & Phòng ban', href: '/branches', icon: Building2, desc: 'Cấu trúc tổ chức hệ thống' },
  { label: 'Nhật ký hệ thống', href: '/audit-logs', icon: ScrollText, desc: 'Toàn bộ audit trail' },
  { label: 'Cài đặt', href: '/settings', icon: Settings, desc: 'Cấu hình hệ thống' },
];

export default function AdminDashboardPage() {
  const { user } = useAuthStore();

  return (
    <div className="max-w-4xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Quản trị hệ thống</h1>
        <p className="text-sm text-text-secondary mt-1">Xin chào, {user?.full_name}. Bạn đang đăng nhập với quyền System Admin.</p>
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

      <div className="rounded-card bg-white border border-border p-4">
        <div className="flex items-center gap-2 mb-3">
          <ShieldCheck className="w-4 h-4 text-primary" />
          <h2 className="text-sm font-semibold text-text-primary">Truy cập toàn bộ hệ thống</h2>
        </div>
        <p className="text-xs text-text-secondary">
          Với quyền SUPER_ADMIN, bạn có thể truy cập tất cả module từ sidebar bên trái.
          Hãy sử dụng thận trọng — mọi thao tác đều được ghi vào audit log.
        </p>
      </div>
    </div>
  );
}
