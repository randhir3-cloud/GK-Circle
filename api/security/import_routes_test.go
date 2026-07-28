package security

import "testing"

func TestCSVImportRoutesInventory(t *testing.T) {
	routes := CSVImportRoutes()
	if len(routes) != 4 {
		t.Fatalf("expected 4 import routes, got %d", len(routes))
	}

	previewCount := 0
	commitCount := 0
	for _, route := range routes {
		if route.Method == "" || route.Path == "" || route.Name == "" {
			t.Fatalf("invalid route inventory entry: %#v", route)
		}
		if !route.RequiresAuth || !route.RequiresQuizEdit {
			t.Fatalf("import route must require auth and quiz edit access: %#v", route)
		}
		if route.UsesPreviewJob && route.PersistsQuestions {
			if route.Name != "import_job_commit" {
				t.Fatalf("only commit route may both preview and persist: %#v", route)
			}
			commitCount++
		}
		if route.UsesPreviewJob && !route.PersistsQuestions {
			previewCount++
		}
	}

	if previewCount != 2 {
		t.Fatalf("expected 2 preview-only routes, got %d", previewCount)
	}
	if commitCount != 1 {
		t.Fatalf("expected 1 commit route, got %d", commitCount)
	}
}
