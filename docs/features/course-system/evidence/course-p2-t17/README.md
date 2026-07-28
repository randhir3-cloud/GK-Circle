# COURSE-P2-T17 Canonical Evidence

Task: `COURSE-P2-T17 — Playwright LearningItem E2E`

Status: VERIFIED

- Verified on: 2026-07-27
- Starting commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Branch: `chore/ci-verification`
- Frozen acceptance source:
  `docs/development/modules/course-system/phases/phase-02-learning-items.md`
- Production source modified: NO
- Database Migration: NONE
- Breaking Changes: NO
- Production/NUC deployment: NONE

The dedicated Playwright workflow uses the existing Chromium project, approved
local Kratos identity, real Nuxt UI, Go API, and PostgreSQL data. It discovers
an existing four-level CourseNode path, creates a DRAFT LearningItem through
the admin UI, proves learner exclusion, publishes through the UI, verifies the
API-ordered learner list and detail, captures desktop/mobile evidence, verifies
signed-out denial, and deletes the temporary row with a 404 read-back.

The focused T17 run passes 1/1 and the frozen full command passes 11/11 without
retries. The full Vitest run reproduces the inherited 35-failure legacy
baseline; all 63 focused Course/LearningItem tests pass.

Evidence index:

- [Acceptance matrix](acceptance-matrix.md)
- [Technical verification](technical-verification.md)
- [Machine-readable verification](technical-verification.json)
- [Baseline](baseline.md)
- [Exact baseline dirty-file inventory](baseline-status.txt)
- [Source mapping](source-mapping.md)
- [Runtime correlation](runtime-verification.md)
- [Diagnostics](diagnostics.md)
- [Cleanup proof](cleanup.md)
- [Production change audit](production-change-audit.md)
- [Changed files](changed-files.md)
- [Screenshots](screenshots.md)
- [Sanitized command log](commands/course-p2-t17.log)
- [Failure trace](traces/focused-timeout-retry-trace.zip)
- [SHA-256 manifest](hashes/sha256.txt)
