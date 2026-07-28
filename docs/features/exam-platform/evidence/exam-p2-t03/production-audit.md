# EXAM-P2-T03 Production Audit

## Closed findings (from EXAM-P1-T02)

| Finding | Route | Resolution |
|---|---|---|
| Unauthenticated analytics with correct answers | `GET /analytics_board/user` | `Authenticated` + `VerifyPlayedQuizReviewAccess` |
| Unauthenticated played-quiz review | `GET /user_played_quizes/:id` | `Authenticated` + `VerifyPlayedQuizReviewAccess` |
| Unauthenticated final score by UUID | `GET /final_score/user` | `Authenticated` + `VerifyPlayedQuizReviewAccess` |

## Rollback

Revert route middleware changes. No migration rollback.

## Residual risk (EXAM-P2-T04)

Pre-release leak proofs and broader endpoint enumeration tests are deferred to T04.

## Runtime verification

Manual smoke: join live quiz → finish → scoreboard loads with cookies; anonymous curl without session returns 401.
