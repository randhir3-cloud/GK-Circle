# COURSE-P1-T02B Technical Verification

Status: PASSED

Technical verification and architectural acceptance are separate:

- Technical verification requires all repository evidence, ledger checks, and
  hash invariants to pass.
- Architectural acceptance occurs afterward by changing ADR-023 from Proposed
  to Accepted and manually changing T02B from IN_PROGRESS to VERIFIED.

## Environment

- Timestamp window: 2026-07-25T07:21:09.1415178Z to
  2026-07-25T07:21:55.2016968Z
- Working directory: `E:\GK Circle v2`
- Environment: local Windows 10.0.26200, PowerShell 5.1.26100.8875
- Time zone: India Standard Time
- Node: v24.16.0
- npm: 11.13.0
- Git: 2.54.0.windows.1
- Branch: `chore/ci-verification`
- Commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`

## Direct repository evidence

- Existing aggregate root: `api/models/course.go:44` defines `Course`.
- Current UUID convention: `api/models/course.go:89` and the other model
  locations recorded in `commands/technical-verification.log` use
  `uuid.NewUUID()`.
- Existing Course migration:
  `api/database/migrations/20260725110500_create_courses_table.up.sql:3` and
  its down migration at line 3.
- Established backend dependencies: `api/go.mod:3` declares Go 1.23,
  `api/go.mod:8` declares goqu, and `api/go.mod:25` declares sql-migrate.
- A corrected repository search found no CourseNode, CourseSubject, or
  CourseTopic persistence model, table, or migration under `api/` or `app/`.
- The accepted path direction originated in
  `docs/development/modules/course-system/DECISIONS.md`; ADR-023 now owns the
  complete normative contract.

## Results

| Check | Exit code | Duration | Classification |
|---|---:|---:|---|
| Checker JavaScript syntax | 0 | 112 ms | PASSED |
| Ledger/evidence JSON parse | 0 | 120 ms | PASSED |
| `git diff --check` | 0 | 183 ms | PASSED |
| Corrected persistence search | 1 | 71 ms | PASSED_NO_MATCHES |
| Initial no-op status sync | 0 | 1396 ms | PASSED |
| Initial read-only status check | 0 | 1139 ms | PASSED |
| Direct architecture evidence scan | 0 | 281 ms | PASSED |
| Migration evidence scan | 0 | 173 ms | PASSED |
| Accepted-state sync | 0 | 1528 ms | PASSED |
| Accepted-state read-only check | 0 | 1052 ms | PASSED_READ_ONLY |
| Final no-op sync | 0 | 991 ms | PASSED_NO_CHANGES |

Ripgrep exit code 1 is the expected successful no-match result for the
persistence search. A discarded earlier search invocation had a PowerShell
quoting error and is not used as supporting evidence.

The initial ledger digest before and after the no-op sync was
`ee491948d2ea26b51cc2f364ab22dd8ec421e449748104d4f2a5eed4504ffc09`.
The digest before and after `--check` was identical, proving that check made no
files changes within the declared verification scope.

## Technical conclusion

Technical verification passed. The architecture decision is additive, the
ledger correctly uses 108 Phase 1 points, no competing hierarchy persistence
exists, sync is idempotent, and check is read-only. Architectural acceptance
may now occur as a separate governance action.

## Accepted-state integrity

After architectural acceptance and all ledger/handoff narrative updates:

- accepted-state sync updated the three declared generated outputs;
- the recursive 35-file ledger-control digest before and after check was
  `444364b7af0f1bea3f92b75a95b986eae206eb18b866237f128687e14f284dbd`;
- check exited 0 and changed no scoped files; and
- the additional sync exited 0 with “no changes required” and the same digest.

The recursive control manifest covers `docs/development/`,
`docs/architecture/ADR/`, the two changed standards files, the ledger checker,
and root `package.json`. The T02B evidence output directory is excluded to
avoid a self-referential digest; its JSON and checker syntax are validated
separately.
