# HRM Permission Test Matrix
## ERP System — MDH Audit Firm
**Version:** 1.0 | **Based on:** HRM_SPEC_v1.4.md §15 | **Usage:** Drives integration tests and manual QA

---

## Legend

| Symbol | Meaning |
|---|---|
| ALLOW | Full access |
| DENY | 403 Forbidden |
| ALLOW-OWN | Only own records |
| ALLOW-BRANCH | Own branch only (HEAD_OF_BRANCH, HR_STAFF) |
| ALLOW-DEPT | Own department only (PARTNER, AUDIT_MANAGER) |
| CONDITIONAL | Depends on condition (see test scenarios) |
| ALLOW-READ | Read-only, no writes |

**11 roles:** SUPER_ADMIN (SA), CHAIRMAN, CEO, HR_MANAGER (HRM), HR_STAFF (HRS), HEAD_OF_BRANCH (HoB), PARTNER, AUDIT_MANAGER (AM), SENIOR_AUDITOR (SA2), JUNIOR_AUDITOR (JA), ACCOUNTANT (ACC)

---

## Matrix: Organization

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| GET branches | ALLOW | ALLOW | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ |
| PUT branches | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET departments | ALLOW | ALLOW | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ |
| PUT departments | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET branch-departments | ALLOW | ALLOW | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ | ALLOW-READ |
| POST branch-departments | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| DELETE branch-departments | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET org-chart | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |

## Matrix: Employees

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| GET /hrm/employees (list) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | DENY | DENY | DENY | DENY | DENY |
| POST /hrm/employees | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET /hrm/employees/:id | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| PUT /hrm/employees/:id | ALLOW | DENY | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| DELETE /hrm/employees/:id | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET /hrm/employees/:id/sensitive | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| POST /hrm/employees/:id/terminate | ALLOW | DENY | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET /my-profile | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| PUT /my-profile (limited fields) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |

## Matrix: Sensitive Fields (Salary, PII)

| Field | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| base_salary | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| cccd_encrypted (decrypt) | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| mst_ca_nhan_encrypted | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| bank_account_encrypted | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| so_bhxh_encrypted | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| commission_rate | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

## Matrix: Dependents, Salary History, Contracts

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| GET dependents | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | DENY | DENY | DENY | DENY | ALLOW-OWN | ALLOW-OWN |
| POST/PUT dependents | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | ALLOW-OWN | ALLOW-OWN |
| DELETE dependents | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET salary-history | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| POST salary-history | ALLOW | DENY | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET contracts | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | ALLOW-OWN | ALLOW-OWN |
| POST/PUT contracts | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

## Matrix: Certifications + Training

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| GET certifications | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| POST cert (own) | ALLOW | DENY | DENY | ALLOW | ALLOW-BRANCH | DENY | DENY | DENY | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| PUT/DELETE cert | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| GET cert/expiring | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY |
| GET training records | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| POST training record | ALLOW | DENY | DENY | ALLOW | ALLOW-BRANCH | DENY | DENY | DENY | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| GET cpe-summary | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| GET training/cpe-summary (all) | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

## Matrix: Leave + OT

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| POST leave request (self) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| GET my leave requests | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| GET all leave requests | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| APPROVE leave | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| REJECT leave | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| CANCEL own leave (PENDING) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| PUT leave-balance (adjust) | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| POST OT request (self) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| APPROVE OT | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| GET OT summary (all) | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

## Matrix: Provisioning + Offboarding

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| POST provisioning request | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | DENY | DENY | DENY | DENY | DENY |
| GET provisioning requests | ALLOW | DENY | DENY | ALLOW | DENY | ALLOW-BRANCH | DENY | DENY | DENY | DENY | DENY |
| branch-approve (step 1) | DENY | DENY | DENY | DENY | DENY | ALLOW-BRANCH | DENY | DENY | DENY | DENY | DENY |
| hr-approve (step 2) | DENY | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| execute (create account) | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| cancel request | ALLOW | DENY | DENY | CONDITIONAL | DENY | CONDITIONAL | DENY | DENY | DENY | DENY | DENY |
| POST offboarding | ALLOW | DENY | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| PUT offboarding item | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

## Matrix: Performance + Independence

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| POST independence declaration | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| GET independence (own) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| GET independence (all) | ALLOW | ALLOW | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| POST acknowledge conflict | DENY | DENY | DENY | DENY | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY |
| POST performance review | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| GET performance reviews | ALLOW | ALLOW | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | ALLOW-OWN | ALLOW-OWN | ALLOW-OWN |
| acknowledge own review | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| POST peer review | ALLOW | DENY | DENY | DENY | DENY | DENY | ALLOW | ALLOW | ALLOW | ALLOW | DENY |

## Matrix: Expenses + Reports

| Action | SA | CHAIRMAN | CEO | HRM | HRS | HoB | PARTNER | AM | SA2 | JA | ACC |
|---|---|---|---|---|---|---|---|---|---|---|---|
| POST expense claim (self) | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| GET own expenses | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW |
| GET all expenses | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| manager-approve (L1) | ALLOW | DENY | ALLOW | ALLOW | DENY | ALLOW-BRANCH | ALLOW-DEPT | ALLOW-DEPT | DENY | DENY | DENY |
| hr-approve (L2) | ALLOW | DENY | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |
| mark-paid | ALLOW | DENY | DENY | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | ALLOW |
| GET HRM reports | ALLOW | ALLOW | ALLOW | ALLOW | ALLOW-BRANCH | ALLOW-BRANCH | DENY | ALLOW-DEPT | DENY | DENY | DENY |
| GET salary report | ALLOW | ALLOW | ALLOW | ALLOW | DENY | DENY | DENY | DENY | DENY | DENY | DENY |

---

## Test Scenarios (BDD Format)

### Organization

```
Scenario: SA creates branch-department mapping
Given: user logged in as SUPER_ADMIN
When: POST /api/v1/hrm/organization/branch-departments {branch_id: HCM, department_id: HR}
Then: 201 Created

Scenario: CEO cannot create branch-department mapping
Given: user logged in as CEO
When: POST /api/v1/hrm/organization/branch-departments
Then: 403 Forbidden
```

### Employees

```
Scenario: HR Manager creates employee
Given: user logged in as HR_MANAGER
When: POST /api/v1/hrm/employees {full_name, branch_id, department_id, grade}
Then: 201 Created, response.data.employee_code matches NV\d{2}-\d{4}

Scenario: Head of Branch reads own-branch employee
Given: user logged in as HEAD_OF_BRANCH for branch=HCM
When: GET /api/v1/hrm/employees/:id (employee.branch_id = HCM)
Then: 200 OK, sensitive fields are "***"

Scenario: Head of Branch cannot read other-branch employee
Given: user logged in as HEAD_OF_BRANCH for branch=HCM
When: GET /api/v1/hrm/employees/:id (employee.branch_id = HO)
Then: 403 Forbidden

Scenario: Partner cannot create employee
Given: user logged in as PARTNER
When: POST /api/v1/hrm/employees
Then: 403 Forbidden

Scenario: Junior Auditor reads own profile only
Given: user logged in as JUNIOR_AUDITOR (employee_id = X)
When: GET /api/v1/hrm/employees/X (own profile)
Then: 200 OK, sensitive fields masked

Scenario: Junior Auditor reads another employee
Given: user logged in as JUNIOR_AUDITOR (employee_id = X)
When: GET /api/v1/hrm/employees/Y (other employee)
Then: 403 Forbidden
```

### PII Access

```
Scenario: HR Manager accesses sensitive PII
Given: user logged in as HR_MANAGER
When: GET /api/v1/hrm/employees/:id/sensitive
Then: 200 OK, decrypted CCCD returned
And: audit_logs contains EMPLOYEE_PII_ACCESSED entry with accessor_id = HR_MANAGER user_id

Scenario: CEO accesses sensitive PII
Given: user logged in as CEO
When: GET /api/v1/hrm/employees/:id/sensitive
Then: 200 OK
And: audit_logs entry created

Scenario: Partner cannot access sensitive PII
Given: user logged in as PARTNER
When: GET /api/v1/hrm/employees/:id/sensitive
Then: 403 Forbidden
And: audit_logs NOT written (no access occurred)

Scenario: HR Staff cannot access sensitive PII
Given: user logged in as HR_STAFF
When: GET /api/v1/hrm/employees/:id/sensitive
Then: 403 Forbidden
```

### Provisioning

```
Scenario: Head of Branch creates provisioning request for own branch
Given: user logged in as HEAD_OF_BRANCH for branch=HCM
When: POST /api/v1/hrm/user-provisioning-requests {employee_id, requested_role: "JUNIOR_AUDITOR", requested_branch_id: HCM}
Then: 201 Created, status = PENDING

Scenario: Head of Branch cannot provision for another branch
Given: user logged in as HEAD_OF_BRANCH for branch=HCM
When: POST /api/v1/hrm/user-provisioning-requests {requested_branch_id: HO}
Then: 403 Forbidden

Scenario: Duplicate PENDING request rejected
Given: employee_id = X has status = PENDING provisioning request
When: POST /api/v1/hrm/user-provisioning-requests {employee_id: X}
Then: 409 Conflict, error = "DUPLICATE_PENDING_REQUEST"

Scenario: SUPER_ADMIN role cannot be provisioned via workflow
Given: any user
When: POST /api/v1/hrm/user-provisioning-requests {requested_role: "SUPER_ADMIN"}
Then: 422 Unprocessable, error = "INVALID_ROLE_FOR_PROVISIONING"

Scenario: Only SA can execute provisioning
Given: user logged in as HR_MANAGER
When: POST /api/v1/hrm/user-provisioning-requests/:id/execute
Then: 403 Forbidden

Scenario: SA executes provisioning atomically
Given: request status = APPROVED, user logged in as SUPER_ADMIN
When: POST /api/v1/hrm/user-provisioning-requests/:id/execute
Then: 200 OK
And: users table has new user record
And: user has requested_role assigned
And: employee.user_id = new user.id
And: audit_logs contains PROVISIONING_EXECUTED
```

### Leave

```
Scenario: Employee submits leave request
Given: user logged in as JUNIOR_AUDITOR with sufficient leave balance
When: POST /api/v1/my-leave/requests {leave_type: "ANNUAL", start_date, end_date, total_days: 2}
Then: 201 Created, status = PENDING
And: leave_balance.pending_days increased by 2

Scenario: Manager approves leave
Given: leave request status = PENDING, user logged in as manager of employee
When: POST /api/v1/hrm/leave/requests/:id/approve
Then: 200 OK, status = APPROVED
And: leave_balance.used_days increased by 2, pending_days decreased by 2 (atomic)

Scenario: Manager rejects leave
Given: leave request status = PENDING
When: POST /api/v1/hrm/leave/requests/:id/reject {rejection_reason: "..."}
Then: 200 OK, status = REJECTED
And: leave_balance.pending_days decreased by 2, used_days unchanged

Scenario: Junior Auditor cannot approve leave
Given: user logged in as JUNIOR_AUDITOR
When: POST /api/v1/hrm/leave/requests/:id/approve
Then: 403 Forbidden

Scenario: Employee can only cancel PENDING requests
Given: leave request status = APPROVED
When: POST /api/v1/my-leave/requests/:id/cancel
Then: 422 Unprocessable (cannot cancel non-PENDING)
```

### OT Cap

```
Scenario: OT approval succeeds within cap
Given: employee has 250 approved OT hours this year
When: Manager approves OT request for 40 hours
Then: 200 OK, total = 290h

Scenario: OT approval at exactly 300h
Given: employee has 290 approved OT hours this year
When: Manager approves OT request for 10 hours
Then: 200 OK, total = 300h

Scenario: OT approval would exceed 300h cap
Given: employee has 290 approved OT hours this year
When: Manager approves OT request for 11 hours
Then: 422 Unprocessable, error = "OT_ANNUAL_CAP_EXCEEDED"

Scenario: OT cap warning at 270h
Given: employee has 260 approved OT hours
When: OT request for 15 hours approved (total = 275h > 270h)
Then: 200 OK
And: audit_logs contains OT_CAP_WARNING entry
And: employee + HR_MANAGER receive notification (event 8)
```

### Independence Declarations

```
Scenario: Employee creates annual declaration (no conflict)
Given: user logged in as any active employee
When: POST /api/v1/hrm/independence {declaration_type: "ANNUAL", declaration_year: 2026, has_conflict: false}
Then: 201 Created, status = CLEAN
And: No notification sent to PARTNER

Scenario: Duplicate annual declaration rejected
Given: employee already has ANNUAL declaration for year 2026
When: POST /api/v1/hrm/independence {declaration_type: "ANNUAL", declaration_year: 2026}
Then: 409 Conflict (unique constraint)

Scenario: Conflict declaration requires Partner acknowledgment
Given: user logged in as JUNIOR_AUDITOR
When: POST /api/v1/hrm/independence {declaration_type: "PER_ENGAGEMENT", engagement_id: X, has_conflict: true}
Then: 201 Created, status = PENDING
And: PARTNER responsible for engagement receives notification (event 17)

Scenario: Partner acknowledges conflict
Given: independence declaration status = PENDING, has_conflict = true
When: POST /api/v1/hrm/independence/:id/acknowledge (PARTNER)
Then: 200 OK, status = CONFLICT_RESOLVED

Scenario: Non-PARTNER cannot acknowledge
Given: independence declaration status = PENDING
When: POST /api/v1/hrm/independence/:id/acknowledge (JUNIOR_AUDITOR)
Then: 403 Forbidden
```

### Expense Claims

```
Scenario: Employee creates expense claim
Given: user logged in as any employee
When: POST /api/v1/my-expenses {claim_period_from, claim_period_to}
Then: 201 Created, status = DRAFT, claim_number matches PC\d{2}-\d{4}

Scenario: Only claim owner can edit DRAFT
Given: expense claim status = DRAFT, created by Employee A
When: PUT /api/v1/my-expenses/:id (Employee B)
Then: 403 Forbidden

Scenario: Accountant can mark expense paid
Given: expense claim status = HR_APPROVED, user logged in as ACCOUNTANT
When: POST /api/v1/hrm/expenses/:id/mark-paid {payment_reference: "REF123"}
Then: 200 OK, status = PAID

Scenario: HR_STAFF cannot mark paid (only HR_MANAGER or ACCOUNTANT)
Given: expense claim status = HR_APPROVED, user logged in as HR_STAFF
When: POST /api/v1/hrm/expenses/:id/mark-paid
Then: 403 Forbidden
```

---

## Edge Cases

### Multiple Roles (Which Wins?)
In this system, users have exactly ONE role. If implementation accidentally allows multiple, the MOST PERMISSIVE role wins — but this should not happen. Flag as bug if detected.

### Self-Access Exceptions
- Employee reads own data: use `/my-profile`, `/my-leave`, `/my-expenses` endpoints (no role restriction)
- Employee reads own data via `/hrm/employees/:id`: only if they are the subject (ALLOW-OWN)
- An employee who is also a manager approves subordinate's requests: use manager scoping

### Branch Scope Crossing
```
Scenario: PARTNER in AUDIT dept (HO) tries to see HCM employee
Given: PARTNER user_id assigned to branch=HO, dept=AUDIT
When: GET /api/v1/hrm/employees/:id (employee.branch_id = HCM, dept = TAX)
Then: 403 Forbidden (PARTNER scope = own dept/branch)
```

### Deleted vs Inactive Records
- `status = TERMINATED`: employee still exists, read-only; HR_MANAGER/CEO/SA can view
- `is_deleted = true` (soft delete): excluded from list queries, 404 on direct access for non-SA
- SA can still access soft-deleted records with explicit filter

### Emergency Flag Bypass
```
Scenario: Emergency provisioning bypasses HoB approval
Given: user logged in as SUPER_ADMIN or CEO
When: POST /api/v1/hrm/user-provisioning-requests {is_emergency: true, emergency_reason: "..."}
Then: Account can be created immediately (SA executes without branch-approve/hr-approve)
And: audit_logs contains PROVISIONING_EMERGENCY entry with emergency_reason
And: 403 if a non-SA/CEO user sets is_emergency=true (field is ignored or rejected)
```

### Encrypted Field Access Control
```
Who can decrypt CCCD, MST, BHXH book, bank account:
  ALLOW: SUPER_ADMIN, CHAIRMAN, CEO, HR_MANAGER
  DENY: ALL other roles (return "***")
  Note: Even HR_STAFF cannot decrypt PII (only HR_MANAGER level)
```

### Audit Log Expectations Per Action

| Action | audit_logs entry | actor | entity |
|---|---|---|---|
| PII accessed | EMPLOYEE_PII_ACCESSED | accessor | employee |
| Employee created | EMPLOYEE_CREATED | creator | employee |
| Leave approved | LEAVE_REQUEST_APPROVED | approver | leave_request |
| OT > 270h approved | OT_CAP_WARNING | approver | ot_request |
| Provisioning executed | PROVISIONING_EXECUTED | SA | provisioning_request |
| Emergency provisioning | PROVISIONING_EMERGENCY | SA/CEO | provisioning_request |
| Salary report run | SALARY_REPORT_GENERATED | accessor | — |
| Independence conflict | INDEPENDENCE_DECLARED | employee | independence_declaration |

### Negative Test Cases (Must Return 403 or 404, Never 200)

| Role | Action | Expected |
|---|---|---|
| JUNIOR_AUDITOR | GET /hrm/employees (list all) | 403 |
| HR_STAFF | GET /hrm/employees/:id/sensitive | 403 |
| PARTNER | POST /hrm/employees (create) | 403 |
| HoB HCM | GET HO employee detail | 403 |
| HoB HCM | PUT /hrm/organization/branches/:id | 403 |
| CEO | DELETE /hrm/employees/:id | 403 |
| ACCOUNTANT | POST provisioning request | 403 |
| JUNIOR_AUDITOR | APPROVE leave request | 403 |
| HR_STAFF | EXECUTE provisioning | 403 |
| PARTNER | mark-paid expense | 403 |
| any | GET non-existent employee ID | 404 |
| any | POST /hrm/invalid-endpoint | 404 |
