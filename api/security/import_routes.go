package security

// ImportRoute describes a CSV import workflow HTTP endpoint audited in EXAM-P3-T04.
type ImportRoute struct {
	Name              string
	Method            string
	Path              string
	RequiresAuth      bool
	RequiresQuizEdit  bool
	PersistsQuestions bool
	UsesPreviewJob    bool
}

// CSVImportRoutes lists the Phase 3 import surfaces.
func CSVImportRoutes() []ImportRoute {
	return []ImportRoute{
		{
			Name:              "import_job_create_preview",
			Method:            "POST",
			Path:              "/api/v1/quizzes/:quiz_id/questions/import-jobs",
			RequiresAuth:      true,
			RequiresQuizEdit:  true,
			PersistsQuestions: false,
			UsesPreviewJob:    true,
		},
		{
			Name:              "import_job_get_preview",
			Method:            "GET",
			Path:              "/api/v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id",
			RequiresAuth:      true,
			RequiresQuizEdit:  true,
			PersistsQuestions: false,
			UsesPreviewJob:    true,
		},
		{
			Name:              "import_job_commit",
			Method:            "POST",
			Path:              "/api/v1/quizzes/:quiz_id/questions/import-jobs/:import_job_id/commit",
			RequiresAuth:      true,
			RequiresQuizEdit:  true,
			PersistsQuestions: true,
			UsesPreviewJob:    true,
		},
		{
			Name:              "legacy_csv_upload",
			Method:            "POST",
			Path:              "/api/v1/quizzes/:quiz_id/questions/upload",
			RequiresAuth:      true,
			RequiresQuizEdit:  true,
			PersistsQuestions: true,
			UsesPreviewJob:    false,
		},
	}
}
