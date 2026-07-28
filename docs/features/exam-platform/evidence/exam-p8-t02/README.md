# EXAM-P8-T02 — Learner analytics aggregations & dashboard

## Scope

Read-only learner dashboard over P8-T01 events + attempts. No instructor analytics, exports, AI, or leaderboards.

## Delivered

- Indexes migration `20260728180000_add_learner_analytics_indexes`
- Aggregation service with bulk release visibility, timezone resolution, dual study-time metrics, versioned Redis cache (fail-open)
- Endpoints under `/api/v1/analytics/*`
- Cache invalidation on submit, release, and telemetry inserts (`inserted > 0` only)
- Nuxt page `/analytics` with summary, study time, trends, subjects, activity, timeline
- Learner nav links for Courses + Analytics
- Mobile layout overflow fix (pending smoke re-run)

## Checks

```text
API:
  go test ./services ./controllers/api/v1 ./models ./utils ./middlewares -count=1 → PASS
  go build ./... → PASS

Frontend:
  npm run lint → PASS
  npm test -- --run → PASS (includes learner analytics component tests)
  npm run build → PASS
```

Smoke: `app/tests/e2e/exam-p8-t02-learner-dashboard-smoke.mjs` (requires running stack).

## Breaking Changes

NO

## Database migration status

Applied in current stack:
- `20260728180000_add_learner_analytics_indexes` (up/down tracked in `gorp_migrations`)
- Indexes exist: `idx_attempts_user_submitted`, `idx_attempts_user_quiz_submitted`

Live smoke status (environment):
Re-run result:
- `SMOKE_OK`
- Mobile layout: PASS (`UI_MOBILE_OK`)
- Recent-activity cursor pagination: SKIPPED (no `has_more` for this dataset)
- Withheld-result protection: SKIPPED (no `Result Pending` activity row for this dataset)
