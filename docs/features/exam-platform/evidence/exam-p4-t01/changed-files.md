# EXAM-P4-T01 — Changed Files

## Migrations

- `api/database/migrations/20260727160000_create_question_collections.up.sql`
- `api/database/migrations/20260727160000_create_question_collections.down.sql`

## Backend

- `api/models/question_collection.go` (new)
- `api/models/question_collection_test.go` (new)
- `api/controllers/api/v1/question_collection_controller.go` (new)
- `api/controllers/api/v1/question_collection_controller_test.go` (new)
- `api/pkg/structs/question_collection_requests.go` (new)
- `api/constants/constant.go` (collection route param + error constants)
- `api/routes/main.go` (`setupQuestionCollectionController`)

## Documentation

- `docs/development/modules/exam-platform/phases/exam-p04-collections-test-builder.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
- `docs/features/exam-platform/evidence/exam-p4-t01/` (this pack)
