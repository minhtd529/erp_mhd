Run the following checklist before marking any module or phase as complete:

## 1. Build
```
go build ./...
```
Fix all compile errors before proceeding.

## 2. Tests
```
go test ./internal/... ./pkg/... -count=1
```
All tests must pass. Do not skip or comment out failing tests — fix them.

## 3. Lint
```
golangci-lint run ./...
```
Fix any lint errors. If golangci-lint is not installed, run `go vet ./...` as a minimum.

## 4. Verify patterns
Check that the module just completed satisfies:
- [ ] Every mutation emits `pkg/audit.Logger.Log(...)` with correct module/resource/action
- [ ] Every list endpoint returns `usecase.PaginatedResult[T]`
- [ ] All domain errors are UPPER_SNAKE_CASE sentinels in `domain/errors.go`
- [ ] Handler maps every domain error to the correct HTTP status code
- [ ] No `TODO Phase X.X` comments remain for the current phase

## 5. Update ROADMAP
- Tick all completed `[ ]` checkboxes in `docs/ROADMAP.md` for this module
- Add a `### <Module> Completed (<date>)` section listing what was shipped

## 6. Report
Print a summary in this format:

```
Module: <name>
Phase:  <X.X>
Tests:  <N> passing
Files:  list key files created/modified
ROADMAP: updated ✓
```
