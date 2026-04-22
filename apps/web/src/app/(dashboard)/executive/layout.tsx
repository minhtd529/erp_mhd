'use client';
import { RoleGuard } from '@/components/layout/role-guard';
import { ROLE_GROUPS } from '@/lib/roles';

export default function ExecutiveLayout({ children }: { children: React.ReactNode }) {
  return (
    <RoleGuard allowedRoles={ROLE_GROUPS.executive}>
      {children}
    </RoleGuard>
  );
}
