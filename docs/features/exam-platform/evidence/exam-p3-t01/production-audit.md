# EXAM-P3-T01 Production Audit

## Scope

CSV import job preview API + additive `question_import_jobs` table. No question bank writes on preview.

## Rollback

1. Revert application deploy.
2. Optional: `migrate down` for `20260727140000_create_question_import_jobs` if empty/unused.

## Residual risk

Preview jobs retain answer keys in `preview_json` for authorised editors only (same trust boundary as question list). Job TTL/cleanup deferred.
