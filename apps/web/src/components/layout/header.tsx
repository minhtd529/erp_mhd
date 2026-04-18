'use client';
import { usePathname } from 'next/navigation';
import { Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';

const PAGE_TITLES: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/clients': 'Khách hàng',
  '/employees': 'Nhân viên',
  '/engagements': 'Hợp đồng kiểm toán',
  '/timesheets': 'Chấm công',
  '/billing/invoices': 'Hóa đơn',
  '/billing/payments': 'Thanh toán',
  '/working-papers': 'Hồ sơ kiểm toán',
  '/commissions': 'Hoa hồng',
  '/commissions/my': 'Hoa hồng của tôi',
  '/reports': 'Báo cáo & Thống kê',
  '/settings': 'Cài đặt',
};

function getTitle(pathname: string): string {
  for (const [path, title] of Object.entries(PAGE_TITLES)) {
    if (pathname === path || pathname.startsWith(path + '/')) return title;
  }
  return 'ERP Audit';
}

export function Header() {
  const pathname = usePathname();
  return (
    <header className="h-14 border-b border-border bg-surface flex items-center px-6 gap-4 sticky top-0 z-10">
      <h1 className="text-base font-semibold text-text-primary flex-1">{getTitle(pathname)}</h1>
      <Button variant="ghost" size="icon" className="text-text-secondary">
        <Bell className="h-5 w-5" />
      </Button>
    </header>
  );
}
