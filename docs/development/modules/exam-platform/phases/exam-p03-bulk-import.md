# EXAM-P3 — Bulk MCQ Import (CSV)



* **Status**: IN_PROGRESS

* **Weight**: 8%

* **Started**: 2026-07-27



## Objective



Validated CSV import with preview, row-level errors, and duplicate detection.



## Scope



- MVP: **CSV only**

- Later: XLSX with separate acceptance (not in this phase)



## Planned tasks



| ID | Title | Status | Dependencies | Evidence |

|---|---|---|---|---|

| EXAM-P3-T01 | CSV import job API + validation preview | VERIFIED | EXAM-P2-T01 | `docs/features/exam-platform/evidence/exam-p3-t01/` |

| EXAM-P3-T02 | Nuxt import wizard | VERIFIED | EXAM-P3-T01 | `docs/features/exam-platform/evidence/exam-p3-t02/` |

| EXAM-P3-T03 | Duplicate detection + row error reporting | VERIFIED | EXAM-P3-T01 | `docs/features/exam-platform/evidence/exam-p3-t03/` |

| EXAM-P3-T04 | Import pipeline regression + phase closure evidence | IMPLEMENTED | EXAM-P3-T01, EXAM-P3-T02, EXAM-P3-T03 | `docs/features/exam-platform/evidence/exam-p3-t04/` |



## EXAM-P3-T01 acceptance (frozen)



<!-- TASK:EXAM-P3-T01:ACCEPTANCE:START -->

- [x] `POST /v1/quizzes/:quiz_id/questions/import-jobs` accepts CSV (edit access + ValidateCsv), validates without inserting questions.

- [x] `GET /v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id` returns stored preview job for the quiz.

- [x] Preview payload includes valid rows with ADR-024 answer-authority defaults (`official_answer`, `authoritative_answer`, `answer_review_status=UNREVIEWED`, `revision_number=1`) and per-row error messages.

- [x] File size limit (existing ValidateCsv) and row-count limit (`MaxRows=500`) enforced.

- [x] Shared CSV row validation reused by legacy `/upload` import and preview path; authority applied via `ApplyAnswerAuthority` / `BuildImportPreviewRow`.

- [x] Existing immediate `/upload` import remains available (backward compatible).

- [x] Additive migration `question_import_jobs`; tests + evidence pack.

<!-- TASK:EXAM-P3-T01:ACCEPTANCE:END -->



## Frozen product rules for this phase



- CSV only in P3; XLSX deferred.

- Import must not bypass question versioning / answer authority when questions are later committed (commit is T02+).

- Duplicate detection is EXAM-P3-T03; Nuxt wizard + commit is EXAM-P3-T02.



## EXAM-P3-T02 acceptance (frozen)



<!-- TASK:EXAM-P3-T02:ACCEPTANCE:START -->

- [x] Nuxt `QuizImportWizard`: CSV upload → preview (valid/invalid rows + messages) → commit → result summary.

- [x] `POST /v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit` commits eligible preview rows via `AppendQuestionsToQuiz` (lineage, revision 1, authority, quiz link).

- [x] Valid-row-only commit; all valid rows in one transaction (all-or-nothing).

- [x] Job states `PREVIEWED`/`FAILED` → `COMMITTING` → `COMMITTED` or `FAILED`; idempotent `COMMITTED`; concurrent commit → 409.

- [x] Auth + quiz edit access; job scoped to quiz; authority revalidated from stored preview at commit.

- [x] Migration `20260727150000_alter_question_import_jobs_commit`; tests + evidence.

- [x] Legacy `POST .../upload` retained.

<!-- TASK:EXAM-P3-T02:ACCEPTANCE:END -->



## EXAM-P3-T03 acceptance (frozen)



<!-- TASK:EXAM-P3-T03:ACCEPTANCE:START -->

- [x] Deterministic duplicate fingerprint (stem, type, options, operational answers).

- [x] Preview flags intra-file and quiz duplicates as `kind=duplicate` row errors; duplicates excluded from commit set.

- [x] Commit re-validates against live quiz fingerprints; 409 on duplicate.

- [x] Legacy `/upload` rejects duplicates with row-level messages.

- [x] Wizard shows Duplicate rows separately from Invalid rows.

- [x] Tests + evidence; no migration.

<!-- TASK:EXAM-P3-T03:ACCEPTANCE:END -->



## EXAM-P3-T04 acceptance (frozen)



<!-- TASK:EXAM-P3-T04:ACCEPTANCE:START -->

- [x] Import route inventory documents preview, get, commit, and legacy upload surfaces (auth + quiz edit required).

- [x] Regression suite proves preview creation, authorized preview retrieval, unauthorized/wrong-quiz job access (404), row validation, duplicate classification, authority defaults at commit boundary, legacy duplicate rejection.

- [x] Existing commit regression preserved: idempotent `COMMITTED`, concurrent commit 409, valid-row-only transactional commit via `AppendQuestionsToQuiz` (revision 1 + `question_revisions` via P2 path).

- [x] No parallel question-creation path; T01–T03 behaviour unchanged.

- [x] `go test ./...` and `go build ./...` pass; evidence pack complete.

- [x] EXAM-P3-T04 is the final planned Phase 3 task; phase closure awaits manual review.

<!-- TASK:EXAM-P3-T04:ACCEPTANCE:END -->
