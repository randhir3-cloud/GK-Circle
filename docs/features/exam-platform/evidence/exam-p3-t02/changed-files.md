# EXAM-P3-T02 Changed Files

## Added

- `api/database/migrations/20260727150000_alter_question_import_jobs_commit.up.sql`
- `api/database/migrations/20260727150000_alter_question_import_jobs_commit.down.sql`
- `api/models/question_import_commit_test.go`
- `app/composables/quiz_import_jobs.js`
- `app/components/quiz-manage/QuizImportWizard.vue`
- `app/test/composables/quiz_import_jobs.test.js`
- `app/test/components/quiz-manage/QuizImportWizard.test.js`
- `docs/features/exam-platform/evidence/exam-p3-t02/`

## Modified

- `api/models/question_import_job.go` — commit statuses, claim/finalize/fail, `QuestionsFromImportPreviewRows`
- `api/services/question_import.go` — `CommitPreviewJob` transactional commit
- `api/controllers/api/v1/question_controller.go` — `CommitQuestionImportJob`
- `api/controllers/api/v1/question_import_job_controller_test.go` — commit controller tests
- `api/models/question_import_job_test.go` — updated select columns
- `api/routes/main.go` — commit route
- `app/pages/admin/quiz/list-quiz/[quiz_id]/index.vue` — wizard integration
- `docs/development/modules/exam-platform/phases/exam-p03-bulk-import.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
- `docs/development/modules/exam-platform/PRODUCT_ROADMAP.md`
