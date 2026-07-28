# COURSE-P2-T18 Canonical Evidence

Task: `COURSE-P2-T18 — Phase 2 documentation update`

Status: VERIFIED

- Started: 2026-07-27
- Verified: 2026-07-27
- Starting commit: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Branch: `chore/ci-verification`
- Frozen acceptance source: `docs/development/modules/course-system/phases/phase-02-learning-items.md`
- Production source modified by T18: NO
- Database Migration: NONE
- Breaking Changes: NO
- Deployment: NONE

## Purpose

This evidence directory records the audit + documentation-only disposition for
COURSE-P2-T18: ensuring the Phase 2 docs keep `publish_state` vs
CourseNode `status` distinct and clarify node-local vs course-wide navigation.

## Frozen acceptance (verbatim)

- [x] Docs keep `publish_state` vs CourseNode `status` distinct; node-local vs course-wide navigation clarified.
- [x] Ledger sync/check pass after documentation updates.

## Proof sources (already present in repository)

1. Phase 2 LearningItem docs:
   - `CourseNode status` vs `LearningItem publish_state` and node-local vs course-wide ownership are
     explicitly stated in `docs/development/modules/course-system/phases/phase-02-learning-items.md`:
     - `CourseNode status ≠ LearningItem publish_state`
     - `Phase 2 previous/next is node-local only; Phase 5 owns course-wide sequence and Continue Learning.`

2. Architecture naming invariant:
   - `Lifecycle naming: CourseNode uses status. LearningItem uses publish_state. Do not conflate.`
     in `docs/development/modules/course-system/architecture/current.md`.

## Evidence index

- [Acceptance matrix](acceptance-matrix.md)
- [Technical verification](technical-verification.md)
- [Machine-readable verification](technical-verification.json)
- [Baseline](baseline-status.txt + baseline-diff.txt)
- [Changed files](changed-files.md)
- [Diagnostics](diagnostics.md)
- [Production change audit](production-change-audit.md)
- [Command log](commands/course-p2-t18.log)
- [SHA-256 manifest](hashes/)

