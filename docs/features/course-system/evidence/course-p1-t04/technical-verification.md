# COURSE-P1-T04 Technical Verification

T04 implemented read-only CourseNode hierarchy repository queries only.

| Check | Classification | Result |
|---|---|---|
| Focused CourseNode model tests | PASSED | Root, child, and nested hierarchy cases passed |
| Docker Go 1.23 vet/test/race/build | PASSED | Repository backend verification passed |
| Compose configuration | PASSED | Local configuration parsed successfully |
| Recursive CTE PostgreSQL validation | PASSED | Temporary local PostgreSQL returned Root, 2, 10, and 100 in numeric preorder |
| Database migration | NONE | No repository migration created or applied |
| Frontend/API verification | NOT_APPLICABLE | No frontend or HTTP API surface changed |
| Ledger integrity | PASSED | Final sync/check/hash evidence passed |

The first temporary PostgreSQL seed attempt used an overlength fixed-width owner
value and failed before Course or CourseNode rows were inserted. The retry used
a valid value and passed. This was test-fixture correction only.

Production and NUC were not accessed.
