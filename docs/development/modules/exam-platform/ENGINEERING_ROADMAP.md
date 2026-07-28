# Exam Platform — Engineering Roadmap

Audience: maintainers and implementing agents.

## Stack gate

Nuxt 3 / Vue 3 / Go Fiber / PostgreSQL / Kratos. Next.js rewrites are forbidden by `AGENTS.md`.

## Phase mapping

| Product phase | Engineering focus | Ledger |
|---|---|---|
| P1 | ADR-024; audit evidence; course publish API; enroll gates; Nuxt course builder + learner outline | [`phases/exam-p01-course-builder.md`](phases/exam-p01-course-builder.md) |
| P2 | Question revisions; answer authority fields; secure review endpoints; MCQ editor | [`phases/exam-p02-question-bank.md`](phases/exam-p02-question-bank.md) |
| P3 | CSV import jobs; validation preview; duplicate detection | [`phases/exam-p03-bulk-import.md`](phases/exam-p03-bulk-import.md) |
| P4 | Collection entities; visual builder; inline question create → bank + link | [`phases/exam-p04-collections-test-builder.md`](phases/exam-p04-collections-test-builder.md) |
| P5 | `assessment_*` Go models; transactional scoring; idempotent submit | [`phases/exam-p05-attempt-scoring.md`](phases/exam-p05-attempt-scoring.md) |
| P6 | Nuxt player wired to P5 only | [`phases/exam-p06-test-player.md`](phases/exam-p06-test-player.md) |
| P7 | Result/review APIs; release policy | [`phases/exam-p07-results-review.md`](phases/exam-p07-results-review.md) |
| P8 | Attempt-derived learning analytics | [`phases/exam-p08-analytics.md`](phases/exam-p08-analytics.md) |
| P9 | Revision queue composition | [`phases/exam-p09-revision.md`](phases/exam-p09-revision.md) |
| P10 | PYQ filters; three coverage metrics; Playwright full loop | [`phases/exam-p10-pyq-coverage-verify.md`](phases/exam-p10-pyq-coverage-verify.md) |

## Evidence root

`docs/features/exam-platform/evidence/`

## Coordination

- Reuse Course System Course/CourseNode/LearningItem and quiz/question tables.
- Cite ADR-024 from COURSE-P4 binding work; do not fork a second scorer.
- Prefer additive SQL migrations; reconcile `attempt_answers` CASCADE vs RESTRICT in ADR-024 / P5.
