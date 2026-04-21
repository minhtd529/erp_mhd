# Claude Code HRM Playbook
## ERP System — MDH Audit Firm
**Version:** 1.0 | **Based on:** HRM_SPEC_v1.4.md | **Purpose:** How to work with Claude Code during HRM implementation

---

> This playbook governs all Claude Code sessions during the 4–6 week HRM implementation. Follow the session ritual, use the prompt templates, respect the phased execution rule, and report in the standard format. Deviations require explicit user approval.

---

## Session Start Ritual

Every new Claude Code session working on HRM MUST begin with these steps — in order:

```
1. Read docs/HRM_SPEC_v1.4.md          (source of truth, 3,222 lines)
2. Read docs/hrm/IMPLEMENTATION_PLAN.md (current sprint + daily task)
3. Read docs/hrm/TECHNICAL_RULES.md    (hard rules, no exceptions)
4. Run: git status                      (understand current state)
5. Run: git log --oneline -10           (understand recent commits)
6. State: "Current sprint: N, Task: X.Y, Status: [what's done/pending]"
```

**Do NOT start writing code until all 6 steps are complete.**

If the spec has been read in a recent session and is still in context, step 1 can be abbreviated to reading the relevant section only. But when in doubt — re-read.

---

## Phased Execution Rule (CRITICAL)

For every HRM feature, execution follows 4 phases in strict order. **STOP between each phase and wait for user approval before proceeding.**

```
Phase 1: Migration
  → Write .up.sql and .down.sql
  → Run make migrate-lint
  → Run round-trip test
  → STOP — show files created, await user approval

Phase 2: Backend
  → Implement Repository → UseCase → Handler
  → Wire routes in Gin
  → Write unit tests for calculations
  → STOP — show files created, await user approval

Phase 3: Frontend
  → Implement service layer (services/*.ts)
  → Implement UI page/component
  → Test loading/empty/error states
  → STOP — show files created, await user approval

Phase 4: Integration Test
  → Write integration test covering happy path + 1 error path
  → Run: go test ./internal/hrm/... -run TestFeatureName
  → STOP — show test output, declare phase complete
```

**Why phases?** Each phase has different risk profile. Migration mistakes can corrupt data. Backend bugs can be fixed without DB changes. UI bugs don't affect data integrity. User reviews each phase independently.

**Never skip a phase.** If Phase 2 needs a small schema change, go back to Phase 1. If Phase 3 needs a new endpoint, go back to Phase 2.

---

## Prompt Templates

Use these templates verbatim when starting work on each artifact type. They include the right context for Claude Code to understand scope and constraints.

### Template: Implement Migration XXXXX

```
Task: Implement migration {MIGRATION_NUMBER}_{MIGRATION_NAME}

SPEC reference: HRM_SPEC_v1.4.md §{SECTION_NUMBER}
Migration description: {WHAT_IT_DOES}

Files to create:
- apps/api/migrations/{MIGRATION_NUMBER}_{MIGRATION_NAME}.up.sql
- apps/api/migrations/{MIGRATION_NUMBER}_{MIGRATION_NAME}.down.sql

Requirements:
- Follow all rules in docs/hrm/TECHNICAL_RULES.md §1 (Migration Rules)
- All enum columns must have CHECK constraints
- All FK columns must have indexes
- Down migration must be symmetric and safe
- Rollback risk: {LOW/MEDIUM/HIGH} — document in file header

Dependencies: {LIST_PRIOR_MIGRATIONS_REQUIRED}

Do NOT include seed data in this migration (seed data goes in 000026 only).
After writing, run: make migrate-lint && make migrate-up && make migrate-down && make migrate-up
Report results. STOP after Phase 1.
```

### Template: Implement API Endpoint Group {name}

```
Task: Implement API endpoints for {ENDPOINT_GROUP_NAME}

SPEC reference: HRM_SPEC_v1.4.md §{13.X}
Endpoints to implement:
{LIST_OF_METHOD_PATH_ROLE_TUPLES_FROM_SPEC}

Files to create/modify:
- apps/api/internal/hrm/{entity}/repository.go (or repository/{entity}.go)
- apps/api/internal/hrm/{entity}/usecase.go
- apps/api/internal/hrm/{entity}/handler.go
- apps/api/internal/hrm/router.go (add routes)

Requirements:
- Strict Handler → UseCase → Repository layering (TECHNICAL_RULES.md §3.1)
- Permission middleware on all routes (TECHNICAL_RULES.md §6.1)
- Branch scope for HoB/HR_STAFF (TECHNICAL_RULES.md §3.9)
- Audit log on all mutations (TECHNICAL_RULES.md §3.7)
- Error wrapping: fmt.Errorf("context: %w", err) (TECHNICAL_RULES.md §3.4)
- log.Printf before any 500 response (TECHNICAL_RULES.md §3.5)
- No SELECT * in queries (TECHNICAL_RULES.md §3.2)

Error codes to use (from SPEC §13): {LIST_RELEVANT_ERROR_CODES}

Do NOT implement UI in this phase.
STOP after Phase 2 complete.
```

### Template: Implement UI Page {path}

```
Task: Implement UI page {PAGE_PATH}

SPEC reference: HRM_SPEC_v1.4.md §{14.X}
Page description: {WHAT_PAGE_DOES}
Allowed roles: {ROLES_FROM_SPEC}

Files to create:
- apps/web/src/app{PAGE_PATH}/page.tsx
- apps/web/src/services/hrm/{entity}Service.ts (if not exists)
- apps/web/src/components/hrm/{ComponentName}.tsx (if needed)

Requirements (TECHNICAL_RULES.md §4):
- API calls through services/*.ts only
- react-query for data fetching
- react-hook-form + zod for forms
- shadcn/ui components only
- Loading state: skeleton/spinner
- Empty state: Vietnamese message + icon
- Error state: Vietnamese error + retry button
- Success toast: Vietnamese after mutation
- Responsive: works at 375px minimum
- No inline styles
- Tailwind classes only

Vietnamese labels required for all UI text.
STOP after Phase 3 complete.
```

### Template: Add Integration Test for {workflow}

```
Task: Write integration test for {WORKFLOW_NAME}

SPEC reference: HRM_SPEC_v1.4.md §{20.2.X}
Test file: apps/api/internal/hrm/{entity}/{entity}_test.go

Test scenarios to cover:
1. Happy path: {DESCRIBE_HAPPY_PATH}
2. Error path 1: {DESCRIBE_ERROR_SCENARIO}

Requirements (TECHNICAL_RULES.md §8):
- Real test database (no mocks)
- Test name format: Test{Feature}_{Scenario}_{ExpectedBehavior}
- Test data set up in TestMain or per-test setup
- Assert DB state after operations (not just HTTP response)
- Assert audit_logs entries where applicable (SPEC §17)
- Clean up test data after each test (or use transactions that rollback)

Run after writing: go test ./internal/hrm/{entity}/... -v -run {TestName}
Report test output. STOP after Phase 4 complete.
```

### Template: Fix Bug in {area}

```
Task: Fix bug in {HRM_AREA}

Description: {DESCRIBE_BUG_AND_EXPECTED_BEHAVIOR}
Error seen: {EXACT_ERROR_MESSAGE_OR_BEHAVIOR}
Reproduction: {STEPS_TO_REPRODUCE}

Before changing anything:
1. Read the relevant SPEC section: HRM_SPEC_v1.4.md §{SECTION}
2. Confirm what the correct behavior should be per SPEC
3. If SPEC is ambiguous → STOP and ask user before fixing

Constraints:
- Do NOT modify committed migrations (TECHNICAL_RULES.md §1.1)
- Do NOT change API paths (TECHNICAL_RULES.md §5.2)
- Fix must not break existing tests
- Write a regression test after fixing

After fix: run make lint test
Report what was changed and why. STOP.
```

---

## Standard Reporting Format

After completing each phase (or after any significant work), report in this exact format:

```
## Work Completed

**Phase:** {1/2/3/4} — {Migration/Backend/Frontend/Integration Test}
**Sprint:** {N}, Task {X.Y}: {Task description}

### Files Created
- `path/to/file.go` ({N} lines) — {one-line description}
- `path/to/file.sql` ({N} lines) — {one-line description}

### Files Modified
- `path/to/existing.go` (lines {from}–{to} changed) — {what changed}

### Verification

| Check | Status |
|---|---|
| `make migrate-lint` | ✓ PASS / ✗ FAIL |
| Round-trip migrate test | ✓ PASS / ✗ FAIL |
| `go build ./...` | ✓ PASS / ✗ FAIL |
| Unit tests | ✓ {X}/{Y} passing / ✗ {failures} |
| Integration tests | ✓ {X}/{Y} passing / ✗ {failures} |
| Manual verification | {what was tested manually} |

### DoD Checklist
- [x] Item 1
- [x] Item 2
- [ ] Item 3 — PENDING (blocked by {reason})

### Not Done / Pending
- {What remains for next session}
- {Any known issues or concerns}

### Next Step Recommendation
{Suggest the specific next action — e.g., "Ready for Phase 2: Backend. Use Template: Implement API Endpoint Group."}

---
STOP — awaiting user approval to proceed.
```

---

## Anti-Patterns (Claude Code Must NOT Do)

These are explicit prohibitions. If any of these are about to happen, STOP and reconsider.

### Commits and Pushes
- ❌ Do NOT run `git commit` without explicit user instruction
- ❌ Do NOT run `git push` — user pushes manually
- ❌ Do NOT run `git add .` (add files individually by name)

### SPEC Violations
- ❌ Do NOT invent fields or table names not in SPEC §11
- ❌ Do NOT rename SPEC-defined names (e.g., don't rename `cccd_encrypted` to `id_card_encrypted`)
- ❌ Do NOT create API paths not listed in SPEC §13
- ❌ Do NOT invent roles beyond the 11 defined in SPEC §3.5
- ❌ Do NOT add features not in Phase 1 scope (see SPEC §2.2 vs §2.3 Deferred)

### Code Quality
- ❌ Do NOT write business logic in handlers (belongs in UseCase)
- ❌ Do NOT use `SELECT *` in any query
- ❌ Do NOT silence errors — always `log.Printf` before returning 500
- ❌ Do NOT return `nil, err` without wrapping with `fmt.Errorf("context: %w", err)`
- ❌ Do NOT hardcode `HRM_ENCRYPTION_KEY` — always `os.Getenv("HRM_ENCRYPTION_KEY")`
- ❌ Do NOT decrypt PII in SQL — only in application (UseCase) layer
- ❌ Do NOT log decrypted PII values in application logs

### Database
- ❌ Do NOT modify any migration file after it's committed to `main`
- ❌ Do NOT add seed data to migrations 000019–000025 (only 000026)
- ❌ Do NOT skip the round-trip migration test before committing

### Testing
- ❌ Do NOT mock the database in integration tests (use real test DB)
- ❌ Do NOT skip tests — fix failures before proceeding

### UI
- ❌ Do NOT use inline styles in React components
- ❌ Do NOT create custom button/input components that duplicate shadcn/ui ones
- ❌ Do NOT call `axios`/`fetch` directly in components — use services/*.ts

---

## When to Escalate to User

Stop work and notify user immediately when:

1. **SPEC ambiguity:** A SPEC section is unclear or contradicts another section
   - Example: §13.2 says "HoB" can read employees, but §15.2 says "ALLOW-BRANCH" — need to confirm if cross-branch is 403 or empty list

2. **SPEC conflict:** Two sections of SPEC give conflicting requirements
   - Example: §8.3 says HR_MANAGER approves step 2, but §15 table shows HRM for HR approve — same thing? Verify.

3. **Technical blocker:** Implementation as specified is technically impossible or would require breaking existing code
   - Example: "The alter employees migration would conflict with an existing index that we didn't know about"

4. **Breaking change needed:** Fixing a bug or adding a feature requires changing code that other modules depend on
   - Example: "To implement branch scope, I need to change the middleware that CRM also uses"

5. **Security concern identified:** A new vulnerability or privacy issue is discovered during implementation
   - Example: "I notice that error messages are returning stack traces that include SQL"

6. **Deferred feature request found:** A SPEC section references Phase 2+ features that are needed for Phase 1 to work
   - Example: "The independence declaration form references a conflict screening API marked as Phase 2 deferred"

7. **Migration risk too high:** A migration would have irreversible effects on production data
   - Example: "Migration 000020 down would drop all employee extended data — need explicit approval to proceed"

---

## Error Recovery Procedures

### If Migration Fails Mid-Way

```bash
# Step 1: Check what state the DB is in
psql -c "\dt" | grep hrm

# Step 2: If partial migration, manually rollback to last known good state
# DO NOT run make migrate-down if unsure — it could make things worse
# Check with user first

# Step 3: If only dev/staging DB, safe to reset:
docker compose down -v && make migrate-up

# Step 4: If the migration was committed to main — create a new correction migration
# NEVER modify the committed migration
# Document in CHANGELOG.md
```

### If Test Fails After Code Change

```bash
# Step 1: Run failing test in isolation with verbose output
go test ./internal/hrm/... -v -run TestFailingTest

# Step 2: Check if it's the test or the implementation
# Read the test carefully — maybe test expectation is wrong

# Step 3: Fix the implementation (not the test) unless test expectation is provably wrong per SPEC

# Step 4: If fixing breaks other tests, diagnose root cause
# Do NOT comment out or skip other tests to make suite pass

# Step 5: Report to user if cannot fix within reasonable investigation
```

### If Compile Breaks Existing Code

```bash
# Step 1: Identify what changed that broke compilation
go build ./... 2>&1

# Step 2: Check if the change to HRM code modified shared types/interfaces

# Step 3: Fix the compilation error — do NOT comment out existing code

# Step 4: Run full test suite: make test (not just HRM tests)

# Step 5: If the fix requires changing non-HRM code → escalate to user before changing
```

### If DB Schema and Code Are Out of Sync

```bash
# Symptom: runtime errors like "column does not exist" or "relation does not exist"
# Step 1: Check which migration is applied
psql -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# Step 2: Check what migrations exist in code
ls apps/api/migrations/*.up.sql | sort

# Step 3: Apply missing migrations
make migrate-up

# Step 4: If dirty=true, the last migration failed — investigate before proceeding
```

### How to Rollback Cleanly (Dev Only)

```bash
# Rollback one migration at a time
make migrate-down   # reverts to previous version

# Verify what was rolled back
psql -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;"

# Rollback to specific version
migrate -path apps/api/migrations -database $DATABASE_URL down 1
```

---

## Session Close Checklist

Before ending any Claude Code session that involved code changes:

- [ ] All changed files are saved (no unsaved edits in editor)
- [ ] `make lint` passes (Go lint + TypeScript lint)
- [ ] `go test ./internal/hrm/...` passes (or known failures documented)
- [ ] No `TODO/FIXME/HACK` added in production code paths
- [ ] No `console.log()` left in TypeScript files
- [ ] Report written in standard format (see §Standard Reporting Format)
- [ ] Next step clearly stated in report
- [ ] Waiting for user approval (STOP state — not mid-implementation)

---

## Quick Reference: SPEC Section Map

| I need to know about... | Read SPEC §... |
|---|---|
| Branch and department structure | §3 |
| All employee fields | §4 |
| BHXH / insurance rates | §5 |
| Certifications / CPE | §6 |
| Leave types and OT cap | §7 |
| Provisioning workflow | §8 |
| Offboarding checklist | §9 |
| Expense claim flow | §10 |
| Database schema / CREATE TABLE | §11 |
| Migration details (up/down) | §12 |
| API endpoints | §13 |
| UI pages and routes | §14 |
| Who can do what (permissions) | §15 |
| Notification events | §16 |
| Audit log events | §17 |
| Encryption rules | §18 |
| Report definitions | §19 |
| Test requirements | §20 |
| Sprint plan and scope | §21 |
| Bootstrap sequence | §22 |
