# EXAM-P2-T03 — Secure Answer-Key and Review Endpoints

Status: VERIFIED
- Started: 2026-07-27
- Verified: 2026-07-27

## Frozen Acceptance

- [x] `GET /v1/analytics_board/user`, `GET /v1/final_score/user`, and `GET /v1/user_played_quizes/:user_played_quiz_id` require authentication.
- [x] Unauthenticated callers receive HTTP 401 (no answer-key payload).
- [x] Authenticated participants and live session hosts retain review access to their played-quiz data.
- [x] Quiz editors (creator, shared read/write/share, public-quiz admin) retain per-participant review preview.
- [x] Existing admin aggregate endpoints (`/analytics_board/admin`, `/final_score/admin`) unchanged.
- [x] Unit/middleware tests prove unauthenticated rejection and participant allow paths.
- [x] Evidence pack under `docs/features/exam-platform/evidence/exam-p2-t03/`.

## Standards loaded

AGENTS.md, CLAUDE.MD, docs/standards/index.md, ADR-024, Product/Engineering roadmaps, EXAM-P2 phase ledger.

## Task understanding

EXAM-P1-T02 documented three inherited live-quiz HTTP endpoints that returned correct answers without authentication. EXAM-P2-T03 closes that leak by requiring `Authenticated` (guest JWT or Kratos) plus `VerifyPlayedQuizReviewAccess`, which authorises the participant, session host, or quiz editor before any answer-key payload is returned. Admin aggregate review endpoints remain available to Kratos-authenticated hosts/editors.

## Architecture notes

| Layer | Implementation |
|---|---|
| Access context | `UserPlayedQuizModel.GetReviewAccessContext` joins `user_played_quizzes` → `active_quizzes` |
| Authorisation | `HasQuizReviewPreviewAccess` — owner, session host, or quiz editor (creator / shared / public-quiz admin) |
| Middleware | `VerifyPlayedQuizReviewAccess` — reads `user_played_quiz` query or `user_played_quiz_id` path param |
| Routes | Chains `Authenticated` + `VerifyPlayedQuizReviewAccess` on the three user review endpoints |
| Live flows | Join flow already uses `Authenticated` POST; scoreboard/analytics use `credentials: "include"` |

## Migration summary

None.

## API changes (behavioural, additive security)

| Method | Path | Change |
|---|---|---|
| GET | `/v1/analytics_board/user?user_played_quiz=` | Now requires auth + review access (401/403) |
| GET | `/v1/final_score/user?user_played_quiz=` | Now requires auth + review access (401/403) |
| GET | `/v1/user_played_quizes/:user_played_quiz_id` | Now requires auth + review access (401/403) |

Admin endpoints unchanged.

## Checks (2026-07-27)

```text
go test ./models/ -run "PlayedQuizReview|HasQuizReview|CanAccessReview" -count=1 → PASS
go test ./middlewares/ -run "VerifyPlayedQuizReview" -count=1                  → PASS
go build ./...                                                                 → PASS
```

## Compatibility verification

- Live quiz participants with valid session JWT retain scoreboard and per-question review.
- Session hosts retain access to participant review for their session.
- Kratos quiz editors retain `/admin/played_quiz/:id` preview via shared/creator permissions.
- Admin aggregate analytics/scoreboard (`/admin` endpoints) unchanged.
- No schema changes; Nuxt client already sends cookies on review fetches.

## Production impact assessment

| Item | Value |
|---|---|
| Breaking changes | **NO** (clients already authenticated in live flow) |
| Migration required | **NO** |
| Risk | Low — closes answer-key leak; may reject anonymous scraping |

## Out of scope (per authorisation)

EXAM-P2-T04 pre-release leak proofs, PCS attempt/results (P5–P7), CSV import, collections, player.

## Production source modified by EXAM-P2-T03: YES (Go API routes + middleware)
