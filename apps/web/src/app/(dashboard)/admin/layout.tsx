'use client';
import { RoleGuard } from '@/components/layout/role-guard';
import { ROLE_GROUPS } from '@/lib/roles';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <RoleGuard allowedRoles={ROLE_GROUPS.sysAdmin}>
      {children}
    </RoleGuard>
  );
}
