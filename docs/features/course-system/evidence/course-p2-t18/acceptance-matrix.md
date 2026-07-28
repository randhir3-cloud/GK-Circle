# COURSE-P2-T18 Acceptance Matrix

Frozen acceptance criteria (verbatim):

- [x] Docs keep `publish_state` vs CourseNode `status` distinct; node-local vs course-wide navigation clarified.
- [x] Ledger sync/check pass after documentation updates.

## Criterion A
Docs keep `publish_state` vs CourseNode `status` distinct; node-local vs course-wide navigation clarified.

Verified documentation sources:

- `docs/development/modules/course-system/phases/phase-02-learning-items.md`
  - `CourseNode status ≠ LearningItem publish_state`
  - `Phase 2 previous/next is node-local only; Phase 5 owns course-wide sequence and Continue Learning.`
- `docs/development/modules/course-system/architecture/current.md`
  - `Lifecycle naming: CourseNode uses status. LearningItem uses publish_state. Do not conflate.`

## Criterion B
Ledger sync/check pass after documentation updates.

Verification method:

- Run:
  - `npm run course-system:status:sync` (first sync)
  - `npm run course-system:status:sync` (second sync; must be no-op)
  - `npm run course-system:status:check`

