'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/stores/auth';
import { getRoleLandingPage } from '@/lib/roles';

interface RoleGuardProps {
  allowedRoles: string[];
  children: React.ReactNode;
}

/**
 * Blocks access to a route section if the user lacks one of the allowedRoles.
 * On denial, redirects to the user's own landing page instead of a generic 403.
 * Must be used inside the (dashboard) layout (i.e. after auth is confirmed).
 */
export function RoleGuard({ allowedRoles, children }: RoleGuardProps) {
  const router = useRouter();
  const { user } = useAuthStore();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const userRoles: string[] = user?.roles ?? [];
    const allowed = allowedRoles.some(r => userRoles.includes(r));
    if (!allowed) {
      router.replace(getRoleLandingPage(userRoles));
      return;
    }
    setReady(true);
  }, [user, allowedRoles, router]);

  if (!ready) return null;
  return <>{children}</>;
}
