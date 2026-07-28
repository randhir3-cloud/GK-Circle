# COURSE-P2-T12 Canonical Evidence

Task: `COURSE-P2-T12 — Enrollment gate for learner item APIs`

Status: **VERIFIED**

- Ledger Version: 1
- Schema Version: 1
- Verified on: 2026-07-27
- Branch: `chore/ci-verification`
- Governance: **D-006** (Course enrollment persistence)
- Prompt: `COURSE-P2-T12-UNBLOCK-01`
- Database Migration: **YES** (`20260727100000_create_course_enrollments_table`)
- Frontend: YES (learner enroll CTA + composable enrollment transport)
- Breaking Changes: **YES** (learner LearningItem GET now requires enrollment)
- Production touched: No
- NUC touched: No

## Blocker resolution

Prior disposition was **BLOCKED** (no enrollment persistence; Migration frozen NONE).

Classification under Project Completion Policy: **Category A** (missing schema/API) with
prior Category C governance freeze. UNBLOCK-01 + D-006 authorized additive schema.

Authenticated-only Kratos access is **not** treated as enrollment.

## Implemented

- Additive `course_enrollments` table (`course_id`, `user_id`, unique pair)
- Repository helpers: `IsUserEnrolled`, `RequireUserEnrolled`, `EnrollUser`, `UnenrollUser`
- Learner LearningItem GET list/get require enrollment before item delivery
- Documented denial: HTTP 404 + `course enrollment required` (no LearningItem payload)
- Learner self-enrollment: `GET|POST|DELETE /api/v1/learner/courses/:course_id/enrollment`
- Learner UI enroll CTA when denial is enrollment-required
- Model + controller tests for enrollment and unenrolled denial

## Evidence files

- `blocker-resolution.md`
- `enrollment-contract.md`
- `technical-verification.md` / `technical-verification.json`
- `changed-files.md`
- `commands/`
- `hashes/`
