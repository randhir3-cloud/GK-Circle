# EXAM-P8-T01 — Assessment analytics event pipeline

## Scope

Immutable append-only analytics foundation only. No learner/instructor dashboards (T02–T04).

## Delivered

- Migration `20260728160000_create_assessment_analytics_events`
- Model/service/controller for authoritative, server telemetry, and client batch events
- Attempt lifecycle + result release wiring (`ATTEMPT_*`, `ATTEMPT_AUTOSAVED`, `RESULT_VIEWED`, release override/scheduled effective)
- `X-Correlation-ID` middleware on attempt/results/analytics routes
- Frontend buffer composable `useAssessmentAnalytics` / `createAssessmentAnalyticsBuffer`
- Smoke script `app/tests/e2e/exam-p8-t01-analytics-events-smoke.mjs`

## Checks

```text
API:
  go test (controllers/services/models/utils/middlewares) -count=1  → PASS
  go test pkg/structs (via alternate test binary; Windows AV locks structs.test.exe) → PASS
  go build ./... → PASS

Frontend:
  npm run lint → PASS
  npm test -- --run → 253 passed (51 files)
  npm run build → PASS
```

Smoke script: `app/tests/e2e/exam-p8-t01-analytics-events-smoke.mjs` (requires running stack + credentials).

## Breaking Changes

NO — additive migration and new endpoints.

## Database migration status

Pending `gk-circle migrate up` on each environment.
