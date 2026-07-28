# GK Circle Project Index

## Read order

1. `PROJECT_INDEX.md`
2. `SYSTEM_ARCHITECTURE.md`
3. Module `MASTER_PLAN.md`
4. Module `CURRENT_STATUS.md`
5. Module `HANDOFF.md`
6. Active phase file
7. Related decisions and canonical evidence

## Active module

Course System

* **Active Phase**: COURSE-P2 â€” Learning Items and Information Blocks
* **In-progress Task**: None
* **Next Task**: None (Phase 2 complete)
* **Overall Module Completion**: 19.84%
* **Next Action**: Begin Phase 3

## Related documents

* **Architecture**: [current.md](modules/course-system/architecture/current.md) â€” recursive CourseNode tree; `parent_id` authoritative; children derived (**D-004**, **D-005**)
* **Decisions**: [DECISIONS.md](modules/course-system/DECISIONS.md) (**D-004**, **D-005**)
* **Documentation Freeze**: [DOCUMENTATION_FREEZE.md](modules/course-system/DOCUMENTATION_FREEZE.md) â€” DOC-CS-T04 certificate
* **Phases**: [ROADMAP.md](modules/course-system/ROADMAP.md) + [`phases/`](modules/course-system/phases/)
* **Evidence**: `docs/features/course-system/evidence/`
* **LearningItem API**: [course-learning-items-v1.md](../api/course-learning-items-v1.md) â€” canonical admin CRUD and learner published-read examples
* **Roadmap**: [ROADMAP.md](modules/course-system/ROADMAP.md) â€” implementation-order index (task tables live in `phases/`)
* **Status**: [CURRENT_STATUS.md](modules/course-system/CURRENT_STATUS.md) + [HANDOFF.md](modules/course-system/HANDOFF.md)
* **ADR-023**: [canonical course hierarchy model](../architecture/ADR/ADR-023-canonical-course-hierarchy-model.md)
* **Agent rule**: Before implementing any Course hierarchy task, read `architecture/current.md`, D-004, D-005, ADR-023, the relevant phase, and dependency tasks. Prefer canonical sources (DOC-CS-T03).
* **Feature Dependencies**: [DEPENDENCIES.md](DEPENDENCIES.md)
* **Feature Matrix**: [FEATURE_MATRIX.md](FEATURE_MATRIX.md)
* **Risks**: [RISKS.md](RISKS.md)
* **Backlog**: [BACKLOG.md](BACKLOG.md)
* **Canonical ledger evidence**: `docs/features/course-system/evidence/ledger-initialization/`
* **Canonical COURSE-P1-T02 evidence**: `docs/features/course-system/evidence/course-p1-t02/`
* **Canonical COURSE-P1-T02B evidence**: `docs/features/course-system/evidence/course-p1-t02b/`
* **Canonical COURSE-P1-T03 evidence**: `docs/features/course-system/evidence/course-p1-t03/`
* **Canonical COURSE-P1-T04 evidence**: `docs/features/course-system/evidence/course-p1-t04/`
* **Canonical COURSE-P1-T05 evidence**: `docs/features/course-system/evidence/course-p1-t05/`
* **Canonical COURSE-P1-T06 evidence**: `docs/features/course-system/evidence/course-p1-t06/`
* **Canonical COURSE-P1-T07 evidence**: `docs/features/course-system/evidence/course-p1-t07/`
* **Canonical COURSE-P1-T08 evidence**: `docs/features/course-system/evidence/course-p1-t08/`
* **Canonical COURSE-P1-T09 evidence**: `docs/features/course-system/evidence/course-p1-t09/`
* **Canonical COURSE-P1-T10 evidence**: `docs/features/course-system/evidence/course-p1-t10/`
* **Canonical COURSE-P2-T01 evidence**: `docs/features/course-system/evidence/course-p2-t01/`
* **Canonical COURSE-P2-T02 evidence**: `docs/features/course-system/evidence/course-p2-t02/`
* **Canonical COURSE-P2-T03 evidence**: `docs/features/course-system/evidence/course-p2-t03/`
* **Canonical COURSE-P2-T04 evidence**: `docs/features/course-system/evidence/course-p2-t04/`
* **Canonical COURSE-P2-T05 evidence**: `docs/features/course-system/evidence/course-p2-t05/`
* **Canonical COURSE-P2-T06 evidence**: `docs/features/course-system/evidence/course-p2-t06/`
* **Canonical COURSE-P2-T07 evidence**: `docs/features/course-system/evidence/course-p2-t07/`
* **Canonical COURSE-P2-T08 evidence**: `docs/features/course-system/evidence/course-p2-t08/`
* **Canonical COURSE-P2-T09 evidence**: `docs/features/course-system/evidence/course-p2-t09/`
* **Canonical COURSE-P2-T10 evidence**: `docs/features/course-system/evidence/course-p2-t10/`
* **Canonical COURSE-P2-T11 evidence**: `docs/features/course-system/evidence/course-p2-t11/`
* **Canonical COURSE-P2-T13 evidence**: `docs/features/course-system/evidence/course-p2-t13/`
* **Canonical COURSE-P2-T14 evidence**: `docs/features/course-system/evidence/course-p2-t14/`
* **Canonical COURSE-P2-T16 evidence**: `docs/features/course-system/evidence/course-p2-t16/`
* **Canonical COURSE-P2-T17 evidence**: `docs/features/course-system/evidence/course-p2-t17/`
* **Canonical COURSE-P2-T20 evidence**: `docs/features/course-system/evidence/course-p2-t20/`
* **Canonical COURSE-P2-T21 evidence**: `docs/features/course-system/evidence/course-p2-t21/`
* **Canonical COURSE-P2-T24 evidence**: `docs/features/course-system/evidence/course-p2-t24/`
* **Canonical COURSE-P2-T25 evidence**: `docs/features/course-system/evidence/course-p2-t25/`
* **Canonical COURSE-P2-T26 evidence**: `docs/features/course-system/evidence/course-p2-t26/`
* **Canonical hierarchy ADR**: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
