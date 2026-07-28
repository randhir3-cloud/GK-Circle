# COURSE-P2-T15 Enrollment Blocker — RESOLVED

Token: `blocker:COURSE-P2-T12` — **RESOLVED** by D-006 / COURSE-P2-T12 VERIFIED.

Enrollment regressions are now executable against:

- `course_enrollments` persistence
- learner LearningItem GET enrollment gate
- documented denial `course enrollment required` (HTTP 404)
- self-enroll endpoints
- controller suites: `TestLearnerLearningItemRoutesDenyUnenrolled`,
  `TestLearnerCourseEnrollmentSelfServe`, and enrolled delivery paths in
  publish/visibility/prev-next/draft suites

This file retains historical context only; it is not an active blocker.
