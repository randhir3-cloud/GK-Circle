# EXAM-P3-T03 — Duplicate Detection + Row Error Reporting

Status: IMPLEMENTED (awaiting manual review)
- Started: 2026-07-27
- Implemented: 2026-07-27

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P3 phase ledger, EXAM-P3-T01/T02 evidence, Question Bank services (`AppendQuestionsToQuiz`, `registerQuestions`).

## Task understanding

EXAM-P3-T03 adds **deterministic duplicate detection** to the CSV import pipeline **before commit**:

1. Detect duplicates **within the uploaded CSV** (later row loses; first row wins).
2. Detect duplicates **against questions already on the target quiz** (stem + type + options + operational answers).
3. Report duplicates as structured row errors in the preview payload (`kind=duplicate`, optional `duplicate_of_row` / `duplicate_question_id`).
4. Re-check at commit time so questions added after preview cannot slip through.
5. Apply the same rules to legacy `POST .../questions/upload`.

Out of scope: fuzzy/AI matching, automatic merging, XLSX, collections, background workers.

## Frozen acceptance

<!-- TASK:EXAM-P3-T03:ACCEPTANCE:START -->

- [x] Deterministic fingerprint: trimmed stem, question type, canonical options, sorted operational answers.
- [x] Preview job creation flags intra-file and quiz duplicates as row errors; duplicates excluded from `valid_rows`.
- [x] Commit endpoint re-validates against current quiz fingerprints; returns 409 on duplicate.
- [x] Legacy `/upload` rejects duplicate rows with 400 and row messages.
- [x] `QuizImportWizard` shows separate Duplicate rows vs Invalid rows sections.
- [x] Unit + controller tests; no DB migration required (errors stored in existing `preview_json`).
- [x] Does not bypass question creation, versioning, or authority paths.

<!-- TASK:EXAM-P3-T03:ACCEPTANCE:END -->

## Duplicate detection architecture

```text
CSV row validation (T01)
        ↓
Build ImportPreviewRow (+ authority defaults)
        ↓
ApplyImportDuplicateDetection
  • intra-file map (first row wins)
  • quiz fingerprint index from quiz_questions ⨝ questions
        ↓
valid_rows + errors (kind=duplicate) → question_import_jobs
        ↓
Nuxt wizard preview
        ↓
CommitPreviewJob
  • re-run FindImportDuplicateErrors against live quiz index
  • then AppendQuestionsToQuiz (unchanged T02 path)
```

## Detection rules

| Rule | Match criteria | Winner | Error |
|---|---|---|---|
| Intra-file | Same fingerprint in CSV | First row (lowest `row_number`) | `duplicate of row N in this CSV file` |
| Quiz bank | Fingerprint matches existing quiz question | Existing question kept | `duplicate of existing quiz question {id}` |
| Not duplicate | Different stem, options, type, or answers | — | — |

Fingerprint excludes points, media, resource (content identity only). No fuzzy matching.

## Migration summary

**None.** Duplicate metadata is stored in existing `preview_json` errors array.

## API changes

No new endpoints. Behavioural changes:

| Endpoint | Change |
|---|---|
| `POST .../import-jobs` | Preview may include `kind: "duplicate"` errors |
| `POST .../import-jobs/:id/commit` | 409 when quiz now contains a duplicate of a preview row |
| `POST .../questions/upload` | 400 when CSV contains duplicate rows |

## Checks (2026-07-27)

```text
go test ./... -count=1                                              → PASS
go build ./...                                                      → PASS
npm test -- --run quiz_import_jobs.test.js QuizImportWizard.test.js → PASS (8)
npx eslint (T03 changed frontend files) --max-warnings 0            → PASS
```

## Compatibility verification

- Import wizard and commit flow unchanged structurally; duplicate rows simply move to errors.
- T01/T02 APIs backward compatible (additive error fields).
- Question creation still uses `AppendQuestionsToQuiz` only at commit.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** (stricter validation; some rows previously importable may now be flagged) |
| Migration required | **NO** |
| Risk | Low |

## Deferred work

| Item | Task |
|---|---|
| Cross-quiz / global bank duplicate detection | Future phase if required |
| XLSX import | Later |
| Fuzzy duplicate suggestions | Out of scope |

## Production source modified by EXAM-P3-T03: YES
