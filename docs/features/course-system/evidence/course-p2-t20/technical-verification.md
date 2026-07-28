# COURSE-P2-T20 Technical Verification

| Check | Result |
|---|---|
| Attach/list/get on CourseNodes at depths 1..10 | PASSED |
| Root (depth 1) vs deep (depth 10) ownership parity | PASSED |
| Cross-Course create/list → `ErrLearningItemCrossCourse` | PASSED |
| Wrong course/node get → `ErrLearningItemNotFound` (no leak) | PASSED |
| Published list/get on depth-10 leaf; draft → not-found | PASSED |
| Ownership SQL shape has no depth/path/`MAX_DEPTH` predicates | PASSED |
| Local `go test ./models/ -run LearningItemOwnership` | PASSED |
| Local `go test ./...` | PASSED |
| Docker `api-verify go vet ./...` | PASSED |
| Docker `api-verify go test ./...` | PASSED |
| `docker compose config --quiet` | PASSED |
| Ledger sync/check + no-op second sync | PASSED |
| Frontend | NOT_APPLICABLE |
| Database migration | NONE |

## Notes

- LearningItem ownership is `(course_id, course_node_id)` only; hierarchy depth is conceptual.
- No product `MAX_DEPTH` and no depth-based LearningItem filtering.
- Delete LearningItem renumbering and new HTTP contracts were not part of frozen T20.
