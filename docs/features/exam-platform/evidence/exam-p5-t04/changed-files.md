# EXAM-P5-T04 — Changed files

## Migration
- `api/database/migrations/20260728120000_create_assessment_attempt_snapshot_items.up.sql`
- `api/database/migrations/20260728120000_create_assessment_attempt_snapshot_items.down.sql`

## Models
- `api/models/assessment_attempt.go`
- `api/models/assessment_attempt_test.go`
- `api/models/assessment_attempt_snapshot_item.go`
- `api/models/assessment_attempt_snapshot_item_test.go`

## Services
- `api/services/assessment_attempt.go`
- `api/services/assessment_attempt_test.go`
- `api/services/assessment_attempt_autosave_test.go`
- `api/services/assessment_attempt_submit_test.go`
- `api/services/pcs_scorer.go` (attempt-item score input + expected max helpers)

## API surface
- `api/controllers/api/v1/assessment_attempt_controller.go`
- `api/controllers/api/v1/assessment_attempt_controller_test.go`
- `api/pkg/structs/assessment_attempt_requests.go`

## Docs
- `docs/development/modules/exam-platform/phases/exam-p05-attempt-scoring.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
- `docs/features/exam-platform/evidence/exam-p5-t04/`
