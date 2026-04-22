// Central role definitions for the entire ERP system.
// All pages should import from here instead of defining local role arrays.

export const ROLES = {
  SUPER_ADMIN: 'SUPER_ADMIN',
  CHAIRMAN: 'CHAIRMAN',
  CEO: 'CEO',
  FIRM_PARTNER: 'FIRM_PARTNER',
  AUDIT_MANAGER: 'AUDIT_MANAGER',
  AUDIT_STAFF: 'AUDIT_STAFF',
  HR_MANAGER: 'HR_MANAGER',
  HR_STAFF: 'HR_STAFF',
  HEAD_OF_BRANCH: 'HEAD_OF_BRANCH',
  ACCOUNTANT: 'ACCOUNTANT',
  CLIENT_ADMIN: 'CLIENT_ADMIN',
  CLIENT_USER: 'CLIENT_USER',
} as const;

export type Role = (typeof ROLES)[keyof typeof ROLES];

// Role groups — used for layout guards and sidebar section visibility
export const ROLE_GROUPS = {
  sysAdmin:  [ROLES.SUPER_ADMIN],
  executive: [ROLES.CHAIRMAN, ROLES.CEO],
  partner:   [ROLES.FIRM_PARTNER],
  audit:     [ROLES.AUDIT_MANAGER, ROLES.AUDIT_STAFF],
  hr:        [ROLES.HR_MANAGER, ROLES.HR_STAFF, ROLES.HEAD_OF_BRANCH],
  client:    [ROLES.CLIENT_ADMIN, ROLES.CLIENT_USER],
} satisfies Record<string, Role[]>;

// All internal staff (non-client) roles
export const INTERNAL_ROLES: Role[] = [
  ROLES.SUPER_ADMIN,
  ROLES.CHAIRMAN,
  ROLES.CEO,
  ROLES.FIRM_PARTNER,
  ROLES.AUDIT_MANAGER,
  ROLES.AUDIT_STAFF,
  ROLES.HR_MANAGER,
  ROLES.HR_STAFF,
  ROLES.HEAD_OF_BRANCH,
];

// Landing page after login, ordered by priority (first match wins)
const LANDING_RULES: { roles: Role[]; path: string }[] = [
  { roles: ROLE_GROUPS.sysAdmin,  path: '/admin/dashboard' },
  { roles: ROLE_GROUPS.executive, path: '/executive/dashboard' },
  { roles: ROLE_GROUPS.hr,        path: '/hrm/dashboard' },
  { roles: ROLE_GROUPS.client,    path: '/client/portal' },
  { roles: [...ROLE_GROUPS.partner, ...ROLE_GROUPS.audit], path: '/dashboard' },
];

export function getRoleLandingPage(userRoles: string[]): string {
  for (const rule of LANDING_RULES) {
    if (rule.roles.some(r => userRoles.includes(r))) return rule.path;
  }
  return '/dashboard';
}

// Module-level access helpers
export const MODULE_ROLES = {
  // System
  userManagement:   [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER],
  auditLogs:        [ROLES.SUPER_ADMIN],
  settings:         [ROLES.SUPER_ADMIN],

  // CRM
  clients:          INTERNAL_ROLES,

  // Engagement
  engagements:      [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.AUDIT_MANAGER, ROLES.AUDIT_STAFF, ROLES.CHAIRMAN, ROLES.CEO],
  workingPapers:    [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.AUDIT_MANAGER, ROLES.AUDIT_STAFF],

  // Timesheet
  timesheets:       [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.AUDIT_MANAGER, ROLES.AUDIT_STAFF],

  // Billing
  billing:          [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.CHAIRMAN, ROLES.CEO],

  // Commissions
  commissions:      [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.AUDIT_MANAGER, ROLES.AUDIT_STAFF],

  // HRM — organization (write: SA/CHAIRMAN/CEO/HR_MANAGER; read-only view: HEAD_OF_BRANCH)
  hrmOrg:           [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER, ROLES.HEAD_OF_BRANCH],
  hrmOrgWrite:      [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER],
  hrmEmployees:     [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.FIRM_PARTNER, ROLES.HR_MANAGER, ROLES.HR_STAFF, ROLES.HEAD_OF_BRANCH],
  hrmEmployeeWrite: [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER],
  hrmSensitive:     [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER],
  hrmSalary:        [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER],
  hrmSalaryWrite:   [ROLES.SUPER_ADMIN, ROLES.CEO, ROLES.HR_MANAGER],

  // HRM — provisioning (§13.12)
  hrmProvisioningRead:    [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER, ROLES.HEAD_OF_BRANCH],
  hrmProvisioningCreate:  [ROLES.HR_MANAGER, ROLES.HEAD_OF_BRANCH, ROLES.CEO],
  hrmProvisioningExecute: [ROLES.SUPER_ADMIN],
  hrmProvisioningBranch:  [ROLES.HEAD_OF_BRANCH],
  hrmProvisioningHR:      [ROLES.HR_MANAGER],

  // HRM — offboarding (§13.13)
  hrmOffboardingRead:     [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER, ROLES.CEO],
  hrmOffboardingCreate:   [ROLES.HR_MANAGER],
  hrmOffboardingItems:    [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER, ROLES.HR_STAFF, ROLES.ACCOUNTANT],

  // HRM — certifications (§13.10)
  hrmCertWrite:           [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER],
  hrmExpiringCert:        [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER],

  // HRM — training courses catalog (§13.10)
  hrmTrainingCourseRead:  [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER, ROLES.HR_STAFF],
  hrmTrainingCourseWrite: [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER],
  hrmTrainingCourseDelete:[ROLES.SUPER_ADMIN],

  // HRM — CPE requirements (§13.11)
  hrmCPERead:             [ROLES.SUPER_ADMIN, ROLES.CHAIRMAN, ROLES.CEO, ROLES.HR_MANAGER],
  hrmCPEWrite:            [ROLES.SUPER_ADMIN, ROLES.HR_MANAGER],

  // Reports
  reports:          [ROLES.SUPER_ADMIN, ROLES.FIRM_PARTNER, ROLES.AUDIT_MANAGER, ROLES.CHAIRMAN, ROLES.CEO],
} satisfies Record<string, Role[]>;

export function hasAnyRole(userRoles: string[], allowed: string[]): boolean {
  return allowed.some(r => userRoles.includes(r));
}
