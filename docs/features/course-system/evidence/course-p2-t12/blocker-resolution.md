# COURSE-P2-T12 Blocker Resolution

## Prior blocker

No Course enrollment persistence model, helper, relation, or documented
equivalent existed. Migration was frozen as NONE pending governance.

## Classification

- **A — Repository implementation blocker:** missing schema, model, gate, APIs
- **C — Documentation/governance:** Migration NONE until D-006

Not external (Category D).

## Resolution

1. Accepted **D-006** authorizing `course_enrollments` persistence and learner GET gate.
2. Implemented additive migration + model + controller gate + self-enroll endpoints.
3. Updated architecture §21/§23 and cleared `blocker:COURSE-P2-T12` in RISKS.
4. Verified with focused Go tests and learner frontend transport tests.

Historical blocker text remains in git history; this directory is now VERIFIED
evidence, not a permanent BLOCKED disposition.
