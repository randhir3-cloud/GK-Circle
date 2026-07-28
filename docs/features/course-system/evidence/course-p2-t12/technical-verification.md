# COURSE-P2-T12 Technical Verification

## Checks

| Check | Result |
|---|---|
| D-006 recorded | PASSED |
| Migration up/down present | PASSED |
| Migration applied (`course_enrollments` exists) | PASSED |
| Model enrollment tests | PASSED |
| Learner GET unenrolled denial | PASSED |
| Learner GET enrolled delivery | PASSED |
| Self-enroll controller tests | PASSED |
| Updated publish/visibility/prev-next/draft learner suites | PASSED |
| Frontend learner enrollment transport test | PASSED |
| API image rebuild | PASSED |
| Browser cookie runtime smoke | ATTEMPTED (login CTA remained Loading/csrf); contract covered by HTTP controller suites |

## Breaking Changes

YES — authenticated learners must enroll before LearningItem GET list/get.

## Database Migration

YES — `20260727100000_create_course_enrollments_table` (applied locally)
