# COURSE-P1-T05 Technical Verification

T05 implemented only transactional CourseNode branch moves.

| Check | Classification | Result |
|---|---|---|
| Focused CourseNode mutation tests | PASSED | Validation, no-op, reparent, cycle, conflict, rollback, and path-boundary cases passed |
| Docker Go 1.23 vet/test/race/build | PASSED | Repository backend verification passed |
| Compose configuration | PASSED | Local configuration parsed successfully |
| PostgreSQL move transaction | PASSED | Two locked subtree paths rewrote atomically; boundary-control path remained unchanged |
| Database migration | NONE | No migration created or applied |
| Frontend/API verification | NOT_APPLICABLE | No frontend or HTTP API surface changed |
| Ledger integrity | PASSED | Final sync/check/hash evidence passed |

The local PostgreSQL verification moved a root and one child under a different
root at exact position `2`. It updated exactly two subtree paths and did not
rewrite a control path with the same textual prefix plus an extra character.

Production and NUC were not accessed.
