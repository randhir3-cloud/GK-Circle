# EXAM-P5-T01 — Changed files

## Migrations
- `api/database/migrations/20260728100000_alter_assessment_attempts_add_snapshot.up.sql`
- `api/database/migrations/20260728100000_alter_assessment_attempts_add_snapshot.down.sql`
- `api/database/migrations/20260728100100_alter_attempt_answers_restrict.up.sql`
- `api/database/migrations/20260728100100_alter_attempt_answers_restrict.down.sql`

## Models / services / API
- `api/models/assessment_attempt.go`
- `api/models/assessment_attempt_test.go`
- `api/models/attempt_answer.go`
- `api/models/quiz.go` (GetSelfPacedMetaByID)
- `api/services/assessment_attempt.go`
- `api/services/assessment_attempt_test.go`
- `api/controllers/api/v1/assessment_attempt_controller.go`
- `api/controllers/api/v1/assessment_attempt_controller_test.go`
- `api/pkg/structs/assessment_attempt_requests.go`
- `api/constants/constant.go`
- `api/routes/main.go`

## Docs
- `docs/development/modules/exam-platform/phases/exam-p05-attempt-scoring.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
- `docs/development/modules/exam-platform/PRODUCT_ROADMAP.md`
- `docs/features/exam-platform/evidence/exam-p5-t01/`
