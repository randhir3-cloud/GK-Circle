# EXAM-P2-T04 Security Audit

Audit date: 2026-07-27

## Method

1. Route inventory from `api/routes/main.go` and EXAM-P1-T02 gap list.
2. Serializer inspection for `correct_answer`, `answers`, `official_answer`, `authoritative_answer`, `is_correct`.
3. WebSocket `sendSingleQuestion` vs `EventShowScore` payload review.
4. Automated regression tests with nested JSON leak detection.

## Sensitive field deny-list

`answers`, `correct_answer`, `official_answer`, `authoritative_answer`, `is_correct`

## Surfaces audited

| Surface | Pre-release? | Answer keys when unauthorised | Status |
|---|---|---|---|
| Review HTTP (3 endpoints) | N/A (post-play) | Blocked (T03 + T04 proofs) | VERIFIED |
| Live question WebSocket delivery | Yes | Excluded from payload | VERIFIED |
| Live scoreboard WebSocket | No (post-question) | Revealed by design | ACCEPTED |
| Public quiz catalog | Yes | No question payload | VERIFIED |
| Editor question list | No (editor auth) | Requires Kratos + permission | ACCEPTED |
| Question revisions API | No (editor auth) | Requires edit access | ACCEPTED (T01) |
| Admin aggregate analytics | Post-session | Kratos only | DEFERRED |

## Overlap with T03

T03 implemented controls; T04 adds proofs and pre-release payload contract. No duplicate middleware.
