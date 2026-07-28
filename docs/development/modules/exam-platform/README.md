# Exam Platform — Module README

Product-driven PCS examination preparation platform built on the existing
Go/Fiber + Nuxt 3 quiz and course engines.

## Authority

| Document | Role |
|---|---|
| [`PRODUCT_ROADMAP.md`](PRODUCT_ROADMAP.md) | Product phases and outcomes |
| [`ENGINEERING_ROADMAP.md`](ENGINEERING_ROADMAP.md) | Engineering tasks, ADRs, evidence |
| [`phases/`](phases/) | Frozen task ledgers with acceptance |
| [`CURRENT_STATUS.md`](CURRENT_STATUS.md) | Next safe action |
| [`DECISIONS.md`](DECISIONS.md) | Module decision summaries |
| ADR-024 | Domain architecture (exam platform) |

## Stack (non-negotiable)

- Frontend: **Nuxt 3 / Vue 3** (not Next.js)
- Backend: Go Fiber, PostgreSQL, Ory Kratos
- Single assessment engine — extend quizzes/questions/`assessment_*`; no parallel Nest/Prisma engine

## Coordination with Course System

- Course System Phase 2 is VERIFIED.
- Course System Phase 4 (Assessments) remains the course-tree binding programme.
- Exam Platform owns PCS study-loop product delivery (bank, collections, attempts, analytics, revision).
- Both programmes cite ADR-024 for self-paced scoring, collections, and snapshots.

## Start order

1. EXAM-P1-T01 — Short ADR
2. EXAM-P1-T02 — Repository audit evidence
3. EXAM-P1-T03 — Course Builder and Publication
4. EXAM-P2 — Question Bank, Versioning, Security
