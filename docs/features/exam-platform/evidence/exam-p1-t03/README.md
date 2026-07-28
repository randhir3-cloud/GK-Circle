# EXAM-P1-T03 — Course Builder and Publication

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27
- Re-verified: 2026-07-27 (authorisation gate)

## Task understanding

EXAM-P1-T03 delivers a usable **admin Course Builder**, **course publication transitions**, **learner enrollment publish gate**, and **learner outline entry path** — reusing the existing Course System APIs and Nuxt stack per ADR-024. No Question Bank, collections, attempt engine, or analytics work is in scope.

## Frozen Acceptance

- [x] Admin can create a Course and add SUBJECT/TOPIC (and SECTION) nodes via Nuxt Course Builder UI without manual IDs/JSON.
- [x] Admin can transition Course `status` among DRAFT / PUBLISHED / ARCHIVED via API (and Builder controls).
- [x] Learner enrollment is rejected for non-PUBLISHED courses; published courses expose an outline entry path.
- [x] Focused backend tests cover status update and enroll publish gate; frontend lint/tests for builder smoke pass where applicable.
- [x] Evidence pack complete under `docs/features/exam-platform/evidence/exam-p1-t03/`; Production source modified by EXAM-P1-T03: YES (documented).

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, P1 phase ledger.

## Architecture notes

| Layer | Decision |
|---|---|
| Admin | Reuses `/admin/courses` + node tree APIs from Course System |
| Publication | `courses.status` enum; `PATCH` partial update |
| Learner gate | `RequirePublishedCourse` + `ErrCourseNotPublished` on enroll |
| Learner UX | `/courses` catalog → `/courses/:id` outline → node learning items when enrolled |
| Out of scope | Question versioning, collections, self-paced attempts |

## API changes (additive)

| Method | Path | Purpose |
|---|---|---|
| `PATCH` | `/api/v1/admin/courses/:course_id` | Accepts `status` (`DRAFT` \| `PUBLISHED` \| `ARCHIVED`) |
| `GET` | `/api/v1/learner/courses` | Published course catalog |
| `GET` | `/api/v1/learner/courses/:course_id` | Published course detail |
| `GET` | `/api/v1/learner/courses/:course_id/nodes/tree` | Learner outline |
| `POST` | `/api/v1/learner/courses/:course_id/enrollment` | Enroll (403 if not published) |

## Frontend surfaces

| Route | Purpose |
|---|---|
| `/admin/courses` | Course Builder (create, nodes, publish) |
| `/courses` | Learner published catalog |
| `/courses/:course_id` | Outline + enrol |

## Checks (2026-07-27)

| Check | Result |
|---|---|
| `go test ./models/ -run "Course|Enrollment" -count=1` | PASS |
| `go test ./controllers/api/v1/ -run "Course|Enrollment" -count=1` | PASS |
| Vitest `CourseLearningItemsApi` + `LearnerLearningItemsApi` | PASS (6 tests) |
| eslint on T03 frontend files | PASS |

## Migration summary

**None for EXAM-P1-T03.** Uses existing `courses` table and `status` column from Course System P1 migrations.

## Compatibility verification

- Existing admin Course/Node/LearningItem APIs unchanged except `PATCH` now accepts `status`.
- Draft/archived courses hidden from learner catalog (`ListPublishedCourses`).
- Enrollment idempotent; publish gate returns `403` with `course is not published`.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | NO |
| New migration | NO |
| Risk | Low — additive learner routes and publish gate |

## Evidence index

- [changed-files.md](changed-files.md)
- [commands/course-p1-t03.log](commands/course-p1-t03.log)
- [production-audit.md](production-audit.md)

## Production source modified by EXAM-P1-T03: YES

Documented intentional production changes under `api/` and `app/` for publish gates, learner catalog/outline, and Course Builder UI.
