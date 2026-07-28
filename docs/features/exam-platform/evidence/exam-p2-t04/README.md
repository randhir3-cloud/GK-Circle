# EXAM-P2-T04 — Answer-Key Protection Tests (Pre-Release Leak Proofs)

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27

## Task understanding

EXAM-P2-T03 implemented authentication and authorisation on the three inherited live-quiz review HTTP endpoints. EXAM-P2-T04 adds the **security proof layer**: regression tests, route inventory, nested JSON leak detection, and an explicit pre-release live-question delivery contract that excludes answer keys. T04 does not duplicate T03 middleware; it proves the controls work and guards against regressions.

## Frozen acceptance

- [x] Unauthenticated review → 401, no answer-key fields in body.
- [x] Authenticated unauthorised → 403, no answer-key fields.
- [x] Authorised participant / host / editor → review payloads retain answer keys.
- [x] Pre-release WebSocket delivery payload excludes answer keys.
- [x] Nested JSON leak detection helper + route inventory.
- [x] `go test ./...` and `go build ./...` pass.
- [x] Evidence pack complete.

## Security findings and resolved exposure paths

| # | Surface | Audit result | T04 action |
|---|---|---|---|
| 1 | `GET /analytics_board/user` | Secured in T03 | Regression tests: 401/403/200 matrix |
| 2 | `GET /final_score/user` | Secured in T03 | Regression tests: 401/403/200 matrix |
| 3 | `GET /user_played_quizes/:id` | Secured in T03 | Regression tests: 401/403/200 matrix |
| 4 | Live WebSocket `EventSendQuestion` | Delivery payload excluded answers (verified) | Extracted `BuildLiveQuestionDeliveryPayload` + tests |
| 5 | Live WebSocket `EventShowScore` | Answers revealed post-question (by design) | Documented; scoreboard builder tested separately |
| 6 | `GET /quizzes/public` | Metadata only, no questions | Fixture assertion in tests |
| 7 | `GET /quizzes/:id/questions` | Kratos + quiz permission required | Inventory entry; unauthenticated blocked by auth |
| 8 | Admin aggregate review endpoints | Kratos only (unchanged per T03) | Inventory; intentionally deferred session scoping |

## Architecture notes

| Component | Purpose |
|---|---|
| `api/security/answer_key.go` | Sensitive field deny-list; nested JSON walker; test assertions |
| `api/security/route_inventory.go` | Documented review / pre-release / editor surfaces |
| `api/utils/live_quiz_payload.go` | Explicit live delivery vs scoreboard payload builders |
| `answer_key_protection_test.go` | End-to-end regression matrix for review access |
| `played_quiz_review_access_test.go` | Extended with host allow path |

## Migration summary

None.

## API behaviour changes

| Change | Detail |
|---|---|
| Live question delivery | Refactored to `BuildLiveQuestionDeliveryPayload` — **no behavioural change**; answer keys still excluded during delivery |
| Review endpoints | No change (T03 controls preserved) |

## Checks (2026-07-27)

```text
go test ./security/... ./utils/ -run "AnswerKey|Live" -count=1  → PASS (6)
go test ./controllers/api/v1/ -run "AnswerKey" -count=1         → PASS (8)
go test ./... -count=1                                          → PASS
go build ./...                                                  → PASS
```

Frontend: no changes; npm tests not required.

## Compatibility verification

- T01 versioning/authority unchanged.
- T02 MCQ editor unchanged.
- T03 review auth unchanged.
- Live scoring and post-question answer reveal unchanged.
- Participant/host/editor review flows preserved via positive regression tests.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** |
| Migration required | **NO** |
| Risk | Very low (tests + payload extraction refactor) |

## Residual risks (intentionally deferred)

| Risk | Notes | Planned phase |
|---|---|---|
| Admin aggregate endpoints (`/analytics_board/admin`, `/final_score/admin`) accept any Kratos user with `active_quiz_id` | Pre-existing; not in T03/T04 scope | EXAM-P7 / hardening |
| PCS attempt/self-paced payloads | Schema exists; no Go runtime | EXAM-P5 |
| Full HTTP route enumeration fuzzing | T04 covers known surfaces + inventory | Future hardening if needed |

## Out of scope

EXAM-P3+, collections, player, analytics engine, revision queue, PYQ.

## Production source modified by EXAM-P2-T04: YES (security package + live payload refactor)
