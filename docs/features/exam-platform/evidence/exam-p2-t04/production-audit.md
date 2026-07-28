# EXAM-P2-T04 Production Audit

## Deployment

Security test package and live payload refactor only. No migration.

## Rollback

Revert `api/security/*`, `api/utils/live_quiz_payload.go`, and socket controller payload calls.

## Runtime smoke (recommended)

1. Join live quiz → receive question without `answers` in WebSocket payload.
2. After question closes → scoreboard event includes `answers`.
3. Anonymous `curl` to review endpoints → 401, no answer keys in body.
