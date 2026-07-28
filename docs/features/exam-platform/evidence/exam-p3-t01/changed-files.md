# EXAM-P3-T01 Changed Files

## Added

- `api/database/migrations/20260727140000_create_question_import_jobs.up.sql`
- `api/database/migrations/20260727140000_create_question_import_jobs.down.sql`
- `api/models/question_import_job.go`
- `api/models/question_import_job_test.go`
- `api/utils/csv_import_preview.go`
- `api/utils/csv_import_preview_test.go`
- `api/services/question_import.go`
- `api/controllers/api/v1/question_import_job_controller_test.go`
- `docs/features/exam-platform/evidence/exam-p3-t01/`

## Modified

- `api/utils/csv_operation.go` — shared row validator; MaxRows on legacy extract
- `api/controllers/api/v1/question_controller.go` — import job create/get handlers
- `api/routes/main.go` — import-jobs routes
- `api/constants/constant.go` — `ImportJobId`
- `docs/development/modules/exam-platform/phases/exam-p03-bulk-import.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
- `docs/development/modules/exam-platform/PRODUCT_ROADMAP.md`
