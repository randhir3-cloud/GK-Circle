# Phase 7 audit enhancements

## Scope

Implemented the Phase 7 review recommendations for result-release audit search and correlation:

1. Indexes on `quiz_result_release_audit_logs` for `(quiz_id, created_at)` and `(actor_id, created_at)`.
2. `correlation_id` column plus request-header resolution (`X-Correlation-ID` → `X-Request-ID` → generated UUID).
3. Confirmed release-policy / audit event constants remain centralised in `api/pkg/structs/quiz_result_management_requests.go`.

## Migration

- `api/database/migrations/20260728150000_alter_quiz_result_release_audit_logs_add_indexes_and_correlation.up.sql`
- Matching `.down.sql`

## Checks

```text
go test ./... -count=1  → PASS
go build ./...          → PASS
```

## Breaking Changes

NO — additive migration only; existing audit rows keep `correlation_id` NULL until new writes.

## Database migration status

Pending apply on each environment via `gk-circle migrate up`.
