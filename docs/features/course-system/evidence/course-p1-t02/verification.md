# COURSE-P1-T02 Verification

Generated: 2026-07-25T11:39:03+05:30

## Environment

- Working copy: `E:\GK Circle v2`
- Branch: `chore/ci-verification`
- Commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Working tree: dirty before T02; unrelated changes and untracked artifacts preserved
- Host: Windows, PowerShell, Go 1.26.5, Node 24.16.0, npm 11.16.0
- Verification containers: Go 1.23.11 Alpine image, PostgreSQL 15.2
- Docker Engine: 29.5.2
- Docker Compose: 5.1.3

## Results

| Verification | Classification | Result |
|---|---|---|
| Focused Course model tests | PASSED | Create, validation, lookup, and missing-row behavior passed |
| Go vet, full tests, race tests | PASSED | Docker `api-verify` exited 0 |
| T02 isolated migration apply | PASSED | Table, columns, constraints, and index created |
| T02 isolated migration rollback | PASSED | Table and T02 migration records removed |
| T02 isolated migration reapply | PASSED | Table and constraints restored |
| Frontend lint (`app/`) | PASSED | ESLint exited 0 |
| Root lint command | NOT_AVAILABLE | Root package intentionally exposes only ledger scripts |
| `git diff --check` | PASSED | Exit 0 |
| Go module tidy diagnostic | FAILED_BASELINE | Existing `api/go.sum` normalization diff; no files written |
| Full historical rollback | FAILED_BASELINE | Older 2024 down migration is invalid; T02 rollback itself completed |
| Ledger sync | PASSED | Generated status files updated from the manually verified task state |
| Ledger check | PASSED | Read-only checker reported a consistent and unchanged ledger |

The exact command metadata is in `verification.json`. No CourseNode, hierarchy,
API, UI, Phase 2, production, or NUC work was performed.

## Changed implementation files

- `api/go.mod`
- `api/models/course.go`
- `api/models/course_test.go`
- `api/database/migrations/20260725110500_create_courses_table.up.sql`
- `api/database/migrations/20260725110500_create_courses_table.down.sql`
- `scripts/course-system-status.js` (derive Course capability flags from the
  manually verified T02 status and gate T03's next action on conflict resolution)

Ledger and canonical evidence files were updated only for T02.
