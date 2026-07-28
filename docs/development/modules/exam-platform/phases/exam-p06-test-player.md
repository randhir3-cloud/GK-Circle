# EXAM-P6 — Student Test Player

* **Status**: IN_PROGRESS
* **Weight**: 10%
* **Started**: 2026-07-28

## Objective

Nuxt learner player wired only to P5 APIs: instructions, palette, timer, mark-for-review, submit confirmation.

## Planned tasks

| ID | Title | Status |
|---|---|---|
| EXAM-P6-T01 | Instructions + start attempt UI | IMPLEMENTED |
| EXAM-P6-T02 | Question palette + autosave feedback | IMPLEMENTED — VERIFICATION HOLD |
| EXAM-P6-T03 | Timer + submit confirmation + expiry handling | NOT_STARTED |

## EXAM-P6-T01 acceptance (frozen)

<!-- TASK:EXAM-P6-T01:ACCEPTANCE:START -->
- [x] Authenticated learner instructions screen for a self-paced quiz + immutable `snapshot_id` (route/query), wired to real attempt/snapshot APIs — no mocked success payloads.
- [x] Instructions present exam rules without answer keys: quiz title, question count, duration, max attempts, negative marking, and start/resume eligibility.
- [x] Start Attempt calls P5 `POST /quizzes/:quiz_id/attempts` with `snapshot_id`; identity from session only; idempotent reuse of an owned `IN_PROGRESS` attempt is supported.
- [x] When an owned `IN_PROGRESS` attempt exists, Resume is offered and navigates into the attempt route using P5 list/get/resume contracts.
- [x] After start/resume, learner lands on a minimal attempt shell (no question palette, autosave UI, timer countdown, or submit confirmation — those remain T02/T03).
- [x] Unauthenticated callers are rejected by the API; learner UI never renders answer keys.
- [x] Analytics, answer review, leaderboards, live quiz player, and EXAM-P6-T02+ remain out of scope.
- [x] Frontend unit tests cover instructions/start/resume wiring; `go test ./...`, `go build ./...`, and applicable `npm` lint/test/build pass; evidence under `docs/features/exam-platform/evidence/exam-p6-t01/`.
<!-- TASK:EXAM-P6-T01:ACCEPTANCE:END -->

## Frozen product rules for this phase

- Player consumes P5 attempt APIs and immutable snapshots (ADR-024); never live bank rows for exam content.
- Scoring remains server-authoritative; the client never invents scores or keys.
- T01 is instructions + start/resume only. T02 adds palette + autosave feedback. T03 adds timer + submit confirmation + expiry.

## Phase-level unresolved verification

- [ ] **Live Compose browser smoke (EXAM-P6)** — end-to-end instructions → start/resume → player → autosave against the real local stack remains an open phase-level verification item. It may stay deferred until Phase 6 closure, but must not be marked complete until executed or explicitly waived at phase close.

## EXAM-P6-T02 acceptance (frozen)

<!-- TASK:EXAM-P6-T02:ACCEPTANCE:START -->
- [x] Attempt player route loads the Phase 5 learner-safe resume payload and renders frozen snapshot questions only (never live bank/collections/editor APIs).
- [x] Explicit player state distinguishes not visited, visited-unanswered, answered, saving, and save failed; answers are not treated as durably saved until autosave succeeds.
- [x] Previous/next navigation, palette navigation, and frozen snapshot ordering are supported; local/saved state is retained across navigation; first/last behaviour is sensible.
- [x] Autosave uses `PUT /v1/quizzes/:quiz_id/attempts/:attempt_id/answers/:question_id` for first save, replacement, clearing, single-answer, survey/multi-select, identical repeats, and retry after failure; client never sends user id, score, correctness, or keys.
- [x] Per-question save sequencing prevents stale responses overwriting newer local selections without blocking unrelated questions.
- [x] Save failures are visible, recoverable, and preserve local draft selections; terminal/authentication/foreign-question errors are handled safely.
- [x] Player UI is keyboard accessible, exposes progress/current question to assistive tech, avoids colour-only palette status, and works at 360 px viewport width.
- [x] Timer, expiry/auto-submit, final submission UX, scoring/results UI, answer review, analytics, leaderboards, and live-player changes remain out of scope.
- [x] Frontend unit tests cover resume rendering, navigation, palette semantics, autosave/retry/stale protection, learner-safe payloads, and T01 compatibility; applicable `npm` lint/test/build pass; evidence under `docs/features/exam-platform/evidence/exam-p6-t02/`.
<!-- TASK:EXAM-P6-T02:ACCEPTANCE:END -->
