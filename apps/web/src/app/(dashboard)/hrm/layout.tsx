'use client';
import { RoleGuard } from '@/components/layout/role-guard';
import { ROLE_GROUPS } from '@/lib/roles';

const HRM_ACCESS = [...ROLE_GROUPS.hr, ...ROLE_GROUPS.sysAdmin, ...ROLE_GROUPS.executive];

export default function HrmLayout({ children }: { children: React.ReactNode }) {
  return (
    <RoleGuard allowedRoles={HRM_ACCESS}>
      {children}
    </RoleGuard>
  );
}
