'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ChevronRight, Bell } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAuthStore } from '@/stores/auth';
import { getModuleContext, MODULE_LABELS, getPageLabel } from '@/lib/navigation';

export function Header() {
  const pathname = usePathname();
  const { user } = useAuthStore();
  const userRoles: string[] = user?.roles ?? [];

  const context = getModuleContext(pathname, userRoles);
  const pageLabel = getPageLabel(pathname);

  return (
    <header className="h-14 border-b border-border bg-surface flex items-center px-6 gap-4 sticky top-0 z-10">
      <div className="flex-1 flex items-center gap-1.5 text-sm min-w-0">
        {context === null ? (
          <span className="font-semibold text-text-primary">Trang chủ</span>
        ) : (
          <>
            <Link href="/" className="text-text-secondary hover:text-primary transition-colors shrink-0">
              Trang chủ
            </Link>
            <ChevronRight className="w-3.5 h-3.5 text-text-secondary/50 shrink-0" />
            <span className="text-text-secondary shrink-0">{MODULE_LABELS[context]}</span>
            {pageLabel && (
              <>
                <ChevronRight className="w-3.5 h-3.5 text-text-secondary/50 shrink-0" />
                <span className="font-semibold text-text-primary truncate">{pageLabel}</span>
              </>
            )}
          </>
        )}
      </div>

      <Button variant="ghost" size="icon" className="text-text-secondary shrink-0">
        <Bell className="h-5 w-5" />
      </Button>
    </header>
  );
}
