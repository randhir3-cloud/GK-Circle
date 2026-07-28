# COURSE-P2-T14 Canonical Evidence

Task: `COURSE-P2-T14 — Learning Item Information Block Rendering`

Status: VERIFIED

- Implemented on: 2026-07-26
- Closure verified: 2026-07-27
- Starting commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Database Migration: NONE
- Backend/API/DTO/persistence changes: NONE
- Production/NUC touched: No
- Breaking Changes: NO

The learner list/detail routes, transport-only learner API composable, immutable
ordered renderer, URL-safety utility, responsive block components, and focused
tests are implemented. The T12 enrollment integration regression was repaired,
the signed-out learner pages were hardened against shrink-to-fit layout, and
all 42 focused T14 tests pass.

The 2026-07-27 closure run completed the authenticated learner audit through the
normal Kratos flow. It correlated PostgreSQL rows with a depth-4 node, learner
responses, and rendered values; proved draft exclusion, empty state,
enrollment-required handling, real enrollment/reload, safe links, backend
adjacency, refresh persistence, boundary omission, signed-out denial, and
desktop/mobile layouts. Temporary representative content was written through
the existing authenticated admin API and fully restored/removed afterward.

The full frontend suite still reproduces the unrelated legacy baseline of 35
failures, now with 96 passing tests and every T14 test green. Lint, Nuxt build,
Go tests, Compose configuration, final `web` image build, service health, status
consistency, evidence integrity, and the scoped diff check pass.

Evidence:

- `technical-verification.md` / `technical-verification.json`
- `changed-files.md`
- `renderer-contract.md`
- `supported-blocks.md`
- `screenshots.md`
- `commands/course-p2-t14.log`
- `hashes/sha256.txt`
- `closure-audit.md`
