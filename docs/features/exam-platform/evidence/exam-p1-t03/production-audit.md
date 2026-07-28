# EXAM-P1-T03 Production Audit

Audit date: 2026-07-27

## Deployment prerequisites

- Course System P1 migrations applied (`courses`, `course_nodes`, `course_enrollments`).
- Ory Kratos authentication configured (admin + learner routes require session).

## Behaviour changes

| Change | Impact |
|---|---|
| `PATCH /admin/courses/:id` accepts `status` | Admins can publish/archive courses |
| Learner catalog/outline routes | New learner entry path for published courses |
| Enrollment publish gate | Draft/archived courses cannot be enrolled |

## Rollback

Revert application deploy; no schema rollback required for T03 alone.

## Security

- Admin routes: Kratos session + admin allowlist (existing Course System pattern).
- Learner routes: Kratos session required.
- Non-published courses return 404 on learner detail/outline (no draft leak via catalog).

## Out of scope risks (unchanged)

Inherited live-quiz unauthenticated review endpoints documented in EXAM-P1-T02 remain open for EXAM-P2-T03 / EXAM-P7.
