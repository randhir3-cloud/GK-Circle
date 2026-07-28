# EXAM-P1 — Course Builder and Publication

* **Status**: VERIFIED
* **Weight**: 10%
* **Started**: 2026-07-27
* **Verified**: 2026-07-27

## Objective

Ship a usable admin Course Builder (course → subjects → topics), course publication transitions, enrollment publish gates, and a learner outline entry point.

## Task table

| ID | Title | Points | Status | Size | Est | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| EXAM-P1-T01 | Architecture + Domain Design (ADR-024) | 3 | VERIFIED | S | 2h | `docs/features/exam-platform/evidence/exam-p1-t01/README.md` | — |
| EXAM-P1-T02 | Repository Audit Evidence | 3 | VERIFIED | S | 2h | `docs/features/exam-platform/evidence/exam-p1-t02/README.md` | — |
| EXAM-P1-T03 | Course Builder and Publication | 8 | VERIFIED | M | 6h | `docs/features/exam-platform/evidence/exam-p1-t03/README.md` | EXAM-P1-T01, EXAM-P1-T02 |

Total points: 14

## Task-specific acceptance criteria

### EXAM-P1-T01 — Architecture + Domain Design (ADR-024)
<!-- TASK:EXAM-P1-T01:ACCEPTANCE:START -->
- [x] ADR-024 is Accepted and lists the minimum decision set (scoring paths, STATIC+DYNAMIC collections, attempt snapshots, answer authority model, question versioning in P2, QUIZ_REFERENCE binding, CASCADE vs RESTRICT intent, Nuxt retained).
- [x] ADR remains short and decision-focused (not a long documentation phase).
- [x] Evidence pack exists under `docs/features/exam-platform/evidence/exam-p1-t01/`.
<!-- TASK:EXAM-P1-T01:ACCEPTANCE:END -->

### EXAM-P1-T02 — Repository Audit Evidence
<!-- TASK:EXAM-P1-T02:ACCEPTANCE:START -->
- [x] Capability status matrix is recorded under exam-platform evidence (COMPLETE / PARTIAL / UNSAFE / MISSING / OUT OF SCOPE).
- [x] Gaps against the PCS MVP study loop are listed with file-path references.
- [x] Evidence pack exists under `docs/features/exam-platform/evidence/exam-p1-t02/`.
<!-- TASK:EXAM-P1-T02:ACCEPTANCE:END -->

### EXAM-P1-T03 — Course Builder and Publication
<!-- TASK:EXAM-P1-T03:ACCEPTANCE:START -->
- [x] Admin can create a Course and add SUBJECT/TOPIC (and SECTION) nodes via Nuxt Course Builder UI without manual IDs/JSON.
- [x] Admin can transition Course `status` among DRAFT / PUBLISHED / ARCHIVED via API (and Builder controls).
- [x] Learner enrollment is rejected for non-PUBLISHED courses; published courses expose an outline entry path.
- [x] Focused backend tests cover status update and enroll publish gate; frontend lint/tests for builder smoke pass where applicable.
- [x] Evidence pack complete under `docs/features/exam-platform/evidence/exam-p1-t03/`; Production source modified by EXAM-P1-T03: YES (documented).
<!-- TASK:EXAM-P1-T03:ACCEPTANCE:END -->
