# COURSE-P2-T11 Technical Verification

| Check | Result |
|---|---|
| Model `MoveLearningItems` + subset validation | PASSED |
| UUID-ordered dual-node locks + six-step staging | PASSED |
| Empty noop with real locked counts | PASSED |
| Concurrent conflict → `ErrLearningItemMoveConflict` / HTTP 409 | PASSED |
| Admin HTTP `POST .../learning-items/move` + frozen payload | PASSED |
| gofmt + `git diff --check` (implementation paths) | PASSED |
| Local `go vet ./...` + `go test ./...` | PASSED |
| Docker `api-verify go vet ./...` | PASSED |
| Docker `api-verify go test ./...` | PASSED |
| `docker compose config --quiet` | PASSED |
| Ledger sync/check + no-op second sync | PASSED |
| Frontend | NOT_APPLICABLE |
| Database migration | NONE |

## Notes

- Sibling lock order per node is `ORDER BY position ASC, id ASC`.
- Cross-node ownership is confined to the locked `(course_id, course_node_id)` pairs.
- Learner APIs, enrollment, visibility evaluation, Prev/Next, and frontend remain out of scope.
