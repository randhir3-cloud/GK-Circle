# COURSE-P2-T10 Technical Verification

| Check | Result |
|---|---|
| Model `ReorderLearningItems` + bidirectional set validation | PASSED |
| Two-phase positions → canonical `0..n-1` | PASSED |
| Idempotent noop (no UPDATE when already canonical) | PASSED |
| Concurrent conflict → `ErrLearningItemReorderConflict` / HTTP 409 | PASSED |
| Admin HTTP `POST .../learning-items/reorder` + frozen payload | PASSED |
| gofmt + `git diff --check` (implementation paths) | PASSED |
| Local `go vet ./...` + `go test ./...` | PASSED |
| Docker `api-verify go vet ./...` | PASSED |
| Docker `api-verify go test ./...` | PASSED |
| `docker compose config --quiet` | PASSED |
| Ledger sync/check + no-op second sync | PASSED |
| Frontend | NOT_APPLICABLE |
| Database migration | NONE |

## Notes

- Sibling lock order is `ORDER BY position ASC, id ASC`.
- Ownership is confined to the locked `(course_id, course_node_id)`.
- Learner APIs, move (T11), enrollment, visibility evaluation, and Prev/Next remain out of scope.
