# COURSE-P2-T22 Canonical Evidence

Task: `COURSE-P2-T22 — Previous/Next learner API (node-local)`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Verified on: 2026-07-26
- Branch: `chore/ci-verification`
- Commit at verification: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Parallel-phase decision: D-003
- Repository adjacent API: Yes (`GetAdjacentPublishedLearningItems`)
- HTTP / learner previous-next: Yes
- Publish filtering: Yes (skips draft / unpublished items at database level)
- Course-wide sequence: No (Phase 5)
- Migration: NONE
- Frontend implemented: No
- Production touched: No
- NUC touched: No
- Breaking Changes: YES (response contract changed from flat to wrapped structure)

This directory is the authoritative T22 evidence store:

- `technical-verification.md` / `technical-verification.json`
- `navigation-contract.md`
- `published-navigation.md`
- `changed-files.md`
- `commands/course-p2-t22.log`
- `hashes/`
