'use client';
import { useAuthStore } from '@/stores/auth';
import { Briefcase, FileText, FolderOpen } from 'lucide-react';

const LINKS = [
  { label: 'Hợp đồng dịch vụ', href: '/engagements', icon: Briefcase, desc: 'Các hợp đồng kiểm toán của bạn' },
  { label: 'Hóa đơn', href: '/billing/invoices', icon: FileText, desc: 'Xem và tra cứu hóa đơn' },
  { label: 'Hồ sơ kiểm toán', href: '/working-papers', icon: FolderOpen, desc: 'Tài liệu kiểm toán được chia sẻ' },
];

export default function ClientPortalPage() {
  const { user } = useAuthStore();

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-text-primary">Cổng thông tin khách hàng</h1>
        <p className="text-sm text-text-secondary mt-1">Xin chào, {user?.full_name}. Đây là các dịch vụ dành cho bạn.</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {LINKS.map(({ label, href, icon: Icon, desc }) => (
          <a
            key={href}
            href={href}
            className="flex flex-col items-center gap-3 p-5 rounded-card bg-white border border-border hover:border-primary/40 hover:shadow-sm transition-all text-center"
          >
            <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
              <Icon className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="text-sm font-semibold text-text-primary">{label}</p>
              <p className="text-xs text-text-secondary mt-0.5">{desc}</p>
            </div>
          </a>
        ))}
      </div>

      <p className="text-xs text-text-secondary">
        Nếu cần hỗ trợ, vui lòng liên hệ kiểm toán viên phụ trách của bạn.
      </p>
    </div>
  );
}
