# EXAM-P3-T03 Changed Files

## Added

- `api/models/import_duplicate.go`
- `api/models/import_duplicate_test.go`
- `docs/features/exam-platform/evidence/exam-p3-t03/`

## Modified

- `api/constants/constant.go` — duplicate error messages
- `api/models/question_import_job.go` — `ImportRowError` duplicate metadata; `ErrImportCommitDuplicates`
- `api/models/questions.go` — `ListImportFingerprintIndexByQuizID`
- `api/services/question_import.go` — preview + commit + legacy duplicate filtering
- `api/controllers/api/v1/question_controller.go` — legacy upload + commit 409
- `api/controllers/api/v1/question_import_job_controller_test.go` — duplicate tests
- `app/components/quiz-manage/QuizImportWizard.vue` — duplicate row UI section
- `app/test/components/quiz-manage/QuizImportWizard.test.js`
- `docs/development/modules/exam-platform/phases/exam-p03-bulk-import.md`
- `docs/development/modules/exam-platform/CURRENT_STATUS.md`
