package v1

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

func newLearnerCourseEnrollmentHTTPEnv(t *testing.T, auth fiber.Handler) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl, err := InitLearnerCourseEnrollmentController(goqu.New("postgres", sqlDB), zap.NewNop(), &config.AppConfig{})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	app := fiber.New()
	learner := app.Group("/api/v1/learner")
	learner.Use(auth)
	courses := learner.Group("/courses")
	courses.Get("/:"+constants.CourseId+"/enrollment", ctrl.GetEnrollment)
	courses.Post("/:"+constants.CourseId+"/enrollment", ctrl.Enroll)
	courses.Delete("/:"+constants.CourseId+"/enrollment", ctrl.Unenroll)
	return app, mock
}

func TestLearnerCourseEnrollmentSelfServe(t *testing.T) {
	app, mock := newLearnerCourseEnrollmentHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	userID := nonAdminUser().ID
	enrollmentID := uuid.MustParse("019c0e00-1111-7000-8000-000000000099")
	now := time.Date(2026, time.July, 27, 2, 0, 0, 0, time.UTC)
	path := "/api/v1/learner/courses/" + courseID.String() + "/enrollment"

	mock.ExpectQuery(`SELECT "id" FROM "course_enrollments"`).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	resp, payload := doJSON(t, app, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status=%d payload=%v", resp.StatusCode, payload)
	}
	data, _ := payload["data"].(map[string]any)
	if data["enrolled"] != false {
		t.Fatalf("expected enrolled=false, got %v", data)
	}

	mock.ExpectQuery(`SELECT "status" FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PUBLISHED"))
	mock.ExpectQuery(`SELECT "id" FROM "course_enrollments"`).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec(`INSERT INTO "course_enrollments"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, payload = doJSON(t, app, http.MethodPost, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("enroll status=%d payload=%v", resp.StatusCode, payload)
	}
	data, _ = payload["data"].(map[string]any)
	if data["enrolled"] != true {
		t.Fatalf("expected enrolled=true, got %v", data)
	}

	mock.ExpectQuery(`SELECT "status" FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PUBLISHED"))
	mock.ExpectQuery(`SELECT "id" FROM "course_enrollments"`).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(enrollmentID))
	mock.ExpectQuery(`SELECT "id", "course_id", "user_id", "enrolled_at" FROM "course_enrollments"`).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "user_id", "enrolled_at"}).
			AddRow(enrollmentID, courseID, userID, now))

	resp, payload = doJSON(t, app, http.MethodPost, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("idempotent enroll status=%d payload=%v", resp.StatusCode, payload)
	}

	mock.ExpectExec(`DELETE FROM "course_enrollments"`).
		WithArgs(courseID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	resp, payload = doJSON(t, app, http.MethodDelete, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unenroll status=%d payload=%v", resp.StatusCode, payload)
	}
	data, _ = payload["data"].(map[string]any)
	if data["enrolled"] != false {
		t.Fatalf("expected enrolled=false after delete, got %v", data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestLearnerCourseEnrollmentMissingCourse(t *testing.T) {
	app, mock := newLearnerCourseEnrollmentHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e499")
	path := "/api/v1/learner/courses/" + courseID.String() + "/enrollment"

	mock.ExpectQuery(`SELECT "status" FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}))

	resp, payload := doJSON(t, app, http.MethodPost, path, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status=%d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseNotFound {
		t.Fatalf("message=%v", payload["data"])
	}
	_ = models.ErrCourseNotFound
}

func TestLearnerCourseEnrollmentDraftRejected(t *testing.T) {
	app, mock := newLearnerCourseEnrollmentHTTPEnv(t, injectAuthenticatedUser(nonAdminUser()))
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e4aa")
	path := "/api/v1/learner/courses/" + courseID.String() + "/enrollment"

	mock.ExpectQuery(`SELECT "status" FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("DRAFT"))

	resp, payload := doJSON(t, app, http.MethodPost, path, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status=%d payload=%v", resp.StatusCode, payload)
	}
	if data, _ := payload["data"].(string); data != constants.ErrCourseNotPublished {
		t.Fatalf("message=%v", payload["data"])
	}
}
