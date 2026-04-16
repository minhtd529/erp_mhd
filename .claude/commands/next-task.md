Read docs/ROADMAP.md and identify the next incomplete task (first unchecked `[ ]` item in the current phase).

Then:
1. State clearly which phase and task you are about to start
2. Read any relevant spec files from docs/modules/ or docs/SPEC.md
3. Implement the task following the established patterns:
   - Module structure: domain → repository → usecase → handler
   - Every mutation MUST emit audit log via pkg/audit
   - Every list endpoint MUST use PaginatedResult[T]
   - Errors use UPPER_SNAKE_CASE sentinel values
   - Tests must cover happy path + key error cases
4. Run `make lint test` (or `go build ./... && go test ./...` if make is unavailable)
5. Update the checkbox in docs/ROADMAP.md when done

Do NOT skip steps or move to the next task until the current one has passing tests.
