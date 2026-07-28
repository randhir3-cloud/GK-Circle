# GK Circle Feature Matrix

This document lists GK Circle's active features, owners, module status, and core dependencies.

## Universal Chained Course System Viewer Matrix

| Feature | Status | Depends On | Module / Phase | Owner |
|---|---|---|---|---|
| Unlimited Recursive Course Tree Foundation | IN_PROGRESS | — | Course System (Phase 1) | AI |
| Learning Items and Information Blocks | IN_PROGRESS | COURSE-P1 structure (verified baseline; parallel UI via D-003) | Course System (Phase 2) | AI |
| Resources and Native Content | NOT_STARTED | COURSE-P2 item composition (design gate first) | Course System (Phase 3) | AI |
| Assessments (any-depth bindings) | NOT_STARTED | COURSE-P2 item composition (design gate first) | Course System (Phase 4) | AI |
| Navigation, Unlocking, and Learning Sequence | NOT_STARTED | COURSE-P1 + COURSE-P2 + COURSE-P3 + COURSE-P4 (consume P1 breadcrumbs; do not redefine) | Course System (Phase 5) | AI |
| Recursive Progress and Completion | NOT_STARTED | COURSE-P5 + COURSE-P2 | Course System (Phase 6) | AI |
| Templates and Builder | NOT_STARTED | COURSE-P1 + COURSE-P2 + COURSE-P3 + COURSE-P4 + COURSE-P6 | Course System (Phase 7) | AI |
| Production Hardening | NOT_STARTED | COURSE-P7 | Course System (Phase 8) | AI |

Hierarchy contract: `CourseNode.parent_id` authoritative; children derived; unlimited logical depth (**D-004** / **D-005**). Canonical ownership graph: `modules/course-system/ROADMAP.md` (DOC-CS-T02-R1 freeze).
