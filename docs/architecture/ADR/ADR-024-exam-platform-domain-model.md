# ADR-024: Exam Platform Domain Model (PCS Study Loop)

- ADR Version: 1
- Schema Version: 1
- Status: Accepted
- Proposed: 2026-07-27
- Accepted: 2026-07-27
- Owners: Architecture and Exam Platform
- Decision task: `EXAM-P1-T01`

## Normative language

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY**
are to be interpreted as described in RFC 2119.

## Executive summary

GK Circle delivers PCS exam preparation by extending the existing Go/Nuxt quiz
and course engines. This ADR freezes domain decisions required before Question
Bank, Collections, Attempt/Scoring, and related schema work.

## Out of scope

This ADR does not itself authorize migrations, APIs, or UI beyond what later
EXAM-P* tasks explicitly accept. It does not authorize NestJS, Next.js, Prisma,
or a second assessment engine.

## Context

Course System Phase 2 is verified. Self-paced SQL
(`assessment_attempts`, `attempt_answers`, quiz self-paced columns) exists
without Go consumers. Live quiz scoring is game-oriented. Product roadmap v3
requires a study loop with versioned MCQs, collections, attempt snapshots, and
distinct coverage metrics.

## Decision

### 1. Stack

- The web client MUST remain Nuxt 3 / Vue 3.
- Next.js / React App Router rewrites MUST NOT be introduced for this programme.

### 2. Single engine; scoring mode split

- Practice, topic, subject, sectional, mock, PYQ, and live modes MUST reuse the
  existing quiz/question/session/scoring infrastructure.
- A parallel Nest/Prisma/`TestsService` engine MUST NOT be created.
- LIVE mode MAY retain time/streak game scoring.
- SELF_PACED (PCS) mode MUST use mark-based scoring including configurable
  negative marks, implemented in EXAM-P5.

### 3. Question Collections (STATIC and DYNAMIC)

- A Question Collection MUST support kind `STATIC` (fixed membership) and
  `DYNAMIC` (resolved from filters such as subject, topic, year, difficulty,
  PYQ status).
- Both kinds MAY exist.
- Every test attempt MUST persist a resolved question snapshot (question IDs,
  revision identifiers, and order). Later collection or bank edits MUST NOT
  alter historical results.

### 4. Answer authority model

Questions MUST model answers as:

| Field | Role |
|---|---|
| `officialAnswer` | Documentary source/exam key |
| `authoritativeAnswer` | Current bank answer for new snapshots |
| `answerReviewStatus` | UNREVIEWED / CONFIRMED / DISPUTED / REVISED |
| `answerRevisionReason` | Why authoritative changed |
| `answerRevisionSource` | Evidence of revision |

The scorer MUST use only the answer key frozen into the test or attempt
snapshot. Historical attempts MUST NOT be silently re-scored against live bank
mutations.

### 5. Question versioning

- Question versioning MUST ship in EXAM-P2 (not deferred to results).
- Edits that change stem, options, answers, or explanations MUST create a
  revision (or equivalent immutable version record).

### 6. Course binding (`QUIZ_REFERENCE`)

- Course-tree assessment entry MUST bind via LearningItem / quiz reference
  mechanisms coordinated with COURSE-P4.
- `QUIZ_REFERENCE` MUST gain an explicit validated link (FK or equivalent
  contract) before learner launch is claimed complete.

### 7. Attempt answer FK policy

- Historical attempt data MUST NOT be destroyed by cascade deletion of parent
  attempts or questions when that would erase auditability of results.
- EXAM-P5 MUST migrate `attempt_answers` toward `ON DELETE RESTRICT` (or an
  equivalent retention strategy) unless a later ADR records a justified
  exception for a specific edge.

### 8. Inline Test Builder question create

- Test Builder MAY create a question without leaving the builder.
- That create MUST use the same Question Bank persistence path and MUST link
  the new question to the current test. No parallel question store.

### 9. Import and coverage

- First import release MUST be CSV-validated; XLSX requires separate acceptance.
- Content coverage, learner coverage, and mastery MUST remain three distinct
  metrics.

### 10. Delivery order

- EXAM-P5 (Attempt and Scoring Engine) MUST precede EXAM-P6 (Student Test
  Player) so learners cannot submit attempts before reliable scoring exists.

## Consequences

- Positive: product loop can proceed without stack rewrite; historical integrity
  is protected by snapshots and versioning.
- Negative: existing self-paced CASCADE migration needs a follow-up migration
  in P5; live review endpoints still require EXAM-P2 security hardening.
- Course System COURSE-P4 MUST cite this ADR for binding and scoring mode
  decisions.

## Evidence

- Task: EXAM-P1-T01
- Pack: `docs/features/exam-platform/evidence/exam-p1-t01/`
