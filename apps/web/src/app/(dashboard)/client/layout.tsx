'use client';
import { RoleGuard } from '@/components/layout/role-guard';
import { ROLE_GROUPS } from '@/lib/roles';

export default function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <RoleGuard allowedRoles={ROLE_GROUPS.client}>
      {children}
    </RoleGuard>
  );
}
