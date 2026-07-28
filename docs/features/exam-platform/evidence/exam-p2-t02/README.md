# EXAM-P2-T02 — Admin MCQ Editor (Nuxt)

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27

## Frozen Acceptance

- [x] Shared `McqQuestionEditor` component for MCQ create/edit: stem, options, media, operational answer, and full ADR-024 answer authority fields.
- [x] `QuestionRevisionHistory` displays immutable revision list via T01 `GET .../revisions` API.
- [x] Quiz manage inline create/edit (`QuestionFormCard`) and dedicated question page (`EditQuestion`) use the shared editor — no duplicate editor logic.
- [x] Official and authoritative answer keys can be set independently of the operational correct answer.
- [x] `quiz_questions` composable and `question_authority` utils centralise API transport and authority payload building.
- [x] Vitest coverage for utils, composable, editor, and `EditQuestion` wrapper.
- [x] Evidence pack under `docs/features/exam-platform/evidence/exam-p2-t02/`.

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P2 phase ledger.

## Architecture notes

| Concept | Implementation |
|---|---|
| Shared editor | `McqQuestionEditor.vue` — single source for MCQ form + authority UI |
| Revision history | `QuestionRevisionHistory.vue` — loads via `useQuizQuestionsApi().listRevisions` |
| Inline quiz manage | `QuestionFormCard.vue` — thin chrome wrapper; passes `quizId` / `questionId` in edit mode |
| Full-page edit | `EditQuestion.vue` — delegates save to composable `updateQuestion` |
| Authority payload | `buildAuthorityPayload()` in `utils/question_authority.js` |
| Backward compatibility | No API/schema changes; reuses EXAM-P2-T01 endpoints |

## Migration summary

None. EXAM-P2-T02 is frontend-only.

## API changes

None (consumes existing T01 endpoints).

## Checks (2026-07-27)

```text
npm test -- --run question_authority quiz_questions McqQuestionEditor EditQuestion → PASS (11)
eslint (T02 frontend files)                                                        → PASS
```

## Compatibility verification

- Existing quiz manage flows unchanged (create, inline edit, full-page edit, reorder, delete).
- Authority fields optional on save; defaults preserve T01 behaviour when keys not overridden.
- No changes to live scoring, CSV import, or attempt engine.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | NO |
| Migration required | NO |
| Downtime | None |
| Risk | Low (admin UI only) |

## Out of scope (per authorisation)

EXAM-P2-T03+, CSV import, collections, test builder, player, analytics, revision queue, PYQ filtering.

## Production source modified by EXAM-P2-T02: YES (Nuxt admin UI)
