# EXAM-P6-T02 — Question Palette + Autosave Feedback

Status: IMPLEMENTED  
- Started: 2026-07-28  
- Implemented: 2026-07-28

## Standards loaded

AGENTS.md, CLAUDE.md, docs/standards/index.md (+ frontend/testing/security), ADR-024, Product/Engineering roadmaps, EXAM-P6 ledger, EXAM-P6-T01 evidence, EXAM-P5 attempt APIs.

## Task understanding

EXAM-P6-T02 delivers the **core self-paced player**: resume → frozen question → select/clear → autosave → palette/progress updates. Timer, submit, results, and review remain out of scope.

## Player state architecture

```text
Resume payload
  → hydrate ordered snapshot items + saved answers
  → per-question state:
       visited, savedSelection, draftSelection,
       saveStatus (idle|saving|saved|failed), saveVersion
  → palette status derived:
       not_visited | visited_unanswered | answered | saving | save_failed
```

## Palette semantics

| Status | Meaning |
|---|---|
| not_visited | Never opened |
| visited_unanswered | Opened; no durable saved selection |
| answered | Autosave succeeded with selection |
| saving | In-flight autosave |
| save_failed | Last save failed; draft preserved |

Palette buttons expose text labels to assistive tech; colour is supplementary.

## Autosave and race control

- Endpoint: `PUT …/attempts/:attempt_id/answers/:question_id`
- Body: `{ selected_options, clear }` only
- Per-question promise chain serialises saves for the same question
- `saveVersion` / `inFlightVersion` reject stale responses
- Unrelated questions save concurrently

## Error and retry behaviour

- Failures set `save_failed` and show retry control
- Local draft preserved on failure
- Terminal attempt responses surface alert and disable inputs
- API errors mapped via `getAssessmentAttemptAPIError` (no key logging)

## Accessibility and responsive verification

- Semantic `fieldset`/`legend` for options
- Palette `aria-current="step"` on active question
- Save status in `role="status"` / `aria-live="polite"`
- Focus-visible styles on controls
- `@media (max-width: 360px)` stacks navigation full-width

## Migration summary

None (frontend-only).

## API changes

None (consumes existing P5 resume + autosave).

## Checks (2026-07-28)

```text
npm test -- --run (T02 suites) → PASS (18 tests)
eslint (T02 files, --max-warnings 0) → PASS
npm run build → PASS
go test ./... → omitted (no backend changes)
Full-repo npm run lint → HUNG (no output after 10+ min; stopped); T02-scoped eslint PASS
Full-repo npm test -- --run → 17 failed | 29 passed; all T02 suites PASS;
  failures are pre-existing outside attempt-player scope
```

## Answer-leak verification

- Player loads resume only; no editor/bank routes
- Unit tests assert instructions/player payloads omit key fields
- Options render from learner-safe snapshot items only

## Compatibility verification

- T01 instructions/start/resume preserved
- P5 autosave semantics unchanged
- Live quiz player untouched

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **NO** |
| Risk | Low |

## Deferred work

| Item | Task |
|---|---|
| Timer + submit confirmation + expiry | EXAM-P6-T03 |
| Results / review | EXAM-P7 |
| Analytics / leaderboards | later |

## Previously unverified T01 checks status

- Full-repo `npm run lint`: **diagnosed** — hangs on `app/playwright-report/` generated JS (not in `.gitignore`). Bounded source-tree eslint completes; T02-scoped eslint PASS. Ignore-list fix deferred pending separate approval.
- Live Compose browser smoke: **still open at phase level** (see EXAM-P6 ledger).

## Verification addendum

See `verification-addendum.md` in this folder (failure classification, baseline comparison, lint diagnosis, integration results).

## Stop condition

**EXAM-P6-T03 not started.** Awaiting explicit manual review / final T02 approval.
