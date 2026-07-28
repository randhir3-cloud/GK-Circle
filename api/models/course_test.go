package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
)

func newCourseModelTest(t *testing.T) (*CourseModel, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sql mock: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return InitCourseModel(goqu.New("postgres", sqlDB)), mock
}

func TestCourseModelCreateCourse(t *testing.T) {
	model, mock := newCourseModelTest(t)
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 25, 11, 5, 0, 0, time.UTC)

	mock.ExpectQuery(`INSERT INTO "courses".*RETURNING`).
		WithArgs(
			"Intermediate",
			sqlmock.AnyArg(),
			"English",
			"owner000000000000001",
			"State PCS foundation",
			"DRAFT",
			"PCS Foundation",
			"PRIVATE",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"owner_id",
			"title",
			"short_description",
			"language",
			"difficulty",
			"visibility",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(
			courseID,
			"owner000000000000001",
			"PCS Foundation",
			"State PCS foundation",
			"English",
			"Intermediate",
			"PRIVATE",
			"DRAFT",
			now,
			now,
		))

	course, err := model.CreateCourse(CreateCourseParams{
		OwnerID:          " owner000000000000001 ",
		Title:            " PCS Foundation ",
		ShortDescription: " State PCS foundation ",
		Language:         " English ",
		Difficulty:       " Intermediate ",
		Visibility:       " PRIVATE ",
	})
	if err != nil {
		t.Fatalf("create course: %v", err)
	}
	if course.ID != courseID {
		t.Fatalf("course ID = %s, want %s", course.ID, courseID)
	}
	if course.Status != CourseStatusDraft {
		t.Fatalf("course status = %s, want %s", course.Status, CourseStatusDraft)
	}
	if course.Title != "PCS Foundation" {
		t.Fatalf("course title = %q, want trimmed title", course.Title)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCourseModelCreateCourseValidation(t *testing.T) {
	model, mock := newCourseModelTest(t)

	_, err := model.CreateCourse(CreateCourseParams{Title: "Course"})
	if !errors.Is(err, ErrCourseOwnerRequired) {
		t.Fatalf("missing owner error = %v, want %v", err, ErrCourseOwnerRequired)
	}

	_, err = model.CreateCourse(CreateCourseParams{OwnerID: "owner"})
	if !errors.Is(err, ErrCourseTitleRequired) {
		t.Fatalf("missing title error = %v, want %v", err, ErrCourseTitleRequired)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("validation unexpectedly queried database: %v", err)
	}
}

func TestCourseModelGetCourseByID(t *testing.T) {
	model, mock := newCourseModelTest(t)
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 25, 11, 5, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT "id", "owner_id", "title", "short_description", "language", "difficulty", "visibility", "status", "created_at", "updated_at" FROM "courses" WHERE ("id" = $1) LIMIT $2`,
	)).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"owner_id",
			"title",
			"short_description",
			"language",
			"difficulty",
			"visibility",
			"status",
			"created_at",
			"updated_at",
		}).AddRow(
			courseID,
			"owner000000000000001",
			"PCS Foundation",
			nil,
			nil,
			nil,
			nil,
			"DRAFT",
			now,
			now,
		))

	course, err := model.GetCourseByID(courseID)
	if err != nil {
		t.Fatalf("get course: %v", err)
	}
	if course.ID != courseID || course.OwnerID != "owner000000000000001" {
		t.Fatalf("unexpected course: %+v", course)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestCourseModelGetCourseByIDNotFound(t *testing.T) {
	model, mock := newCourseModelTest(t)
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")

	mock.ExpectQuery(`SELECT .* FROM "courses"`).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"owner_id",
			"title",
			"short_description",
			"language",
			"difficulty",
			"visibility",
			"status",
			"created_at",
			"updated_at",
		}))

	_, err := model.GetCourseByID(courseID)
	if !errors.Is(err, ErrCourseNotFound) {
		t.Fatalf("get missing course error = %v, want %v", err, ErrCourseNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func courseSelectRows(courses ...Course) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"id", "owner_id", "title", "short_description", "language",
		"difficulty", "visibility", "status", "created_at", "updated_at",
	})
	for _, course := range courses {
		var short, language, difficulty, visibility interface{}
		if course.ShortDescription.Valid {
			short = course.ShortDescription.String
		}
		if course.Language.Valid {
			language = course.Language.String
		}
		if course.Difficulty.Valid {
			difficulty = course.Difficulty.String
		}
		if course.Visibility.Valid {
			visibility = course.Visibility.String
		}
		rows.AddRow(
			course.ID, course.OwnerID, course.Title, short, language,
			difficulty, visibility, course.Status, course.CreatedAt, course.UpdatedAt,
		)
	}
	return rows
}

func TestCourseModelListCoursesEmptyAndOrdered(t *testing.T) {
	t.Run("empty non-nil", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		mock.ExpectQuery(`SELECT .* FROM "courses".*ORDER BY`).
			WillReturnRows(courseSelectRows())
		courses, err := model.ListCourses()
		if err != nil {
			t.Fatalf("list courses: %v", err)
		}
		if courses == nil || len(courses) != 0 {
			t.Fatalf("courses = %#v, want non-nil empty slice", courses)
		}
	})

	t.Run("deterministic order", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		newer := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
		older := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
		firstID := uuid.MustParse("019c0146-aaaa-7f74-a925-9c63b8ea5eed")
		secondID := uuid.MustParse("019c0146-bbbb-7f74-a925-9c63b8ea5eed")
		// Same created_at: id ASC places aaaa before bbbb; newer created_at first overall.
		mock.ExpectQuery(`SELECT .* FROM "courses".*ORDER BY`).
			WillReturnRows(courseSelectRows(
				Course{ID: firstID, OwnerID: "owner000000000000001", Title: "Newer", Status: CourseStatusDraft, CreatedAt: newer, UpdatedAt: newer},
				Course{ID: secondID, OwnerID: "owner000000000000001", Title: "Older", Status: CourseStatusDraft, CreatedAt: older, UpdatedAt: older},
			))
		courses, err := model.ListCourses()
		if err != nil {
			t.Fatalf("list courses: %v", err)
		}
		if len(courses) != 2 || courses[0].ID != firstID || courses[1].ID != secondID {
			t.Fatalf("unexpected order: %+v", courses)
		}
	})
}

func TestCourseModelUpdateCourse(t *testing.T) {
	courseID := uuid.MustParse("019c0146-5f3f-7f74-a925-9c63b8ea5eed")
	now := time.Date(2026, time.July, 26, 7, 0, 0, 0, time.UTC)
	title := "Updated Title"

	t.Run("supplied fields update", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
			WithArgs(nil, title, courseID).
			WillReturnRows(courseSelectRows(Course{
				ID: courseID, OwnerID: "owner000000000000001", Title: title,
				Status: CourseStatusDraft, CreatedAt: now, UpdatedAt: now,
			}))
		course, err := model.UpdateCourse(courseID, UpdateCourseParams{
			Title:            &title,
			ShortDescription: OptionalNullableString{Present: true, Null: true},
		})
		if err != nil {
			t.Fatalf("update course: %v", err)
		}
		if course.Title != title || course.ShortDescription.Valid {
			t.Fatalf("unexpected course: %+v", course)
		}
	})

	t.Run("empty patch rejected without query", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		_, err := model.UpdateCourse(courseID, UpdateCourseParams{})
		if !errors.Is(err, ErrCourseUpdateRequired) {
			t.Fatalf("error = %v, want %v", err, ErrCourseUpdateRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("empty update queried database: %v", err)
		}
	})

	t.Run("empty title rejected", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		empty := "  "
		_, err := model.UpdateCourse(courseID, UpdateCourseParams{Title: &empty})
		if !errors.Is(err, ErrCourseTitleRequired) {
			t.Fatalf("error = %v, want %v", err, ErrCourseTitleRequired)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("invalid title queried database: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
			WithArgs(title, courseID).
			WillReturnRows(courseSelectRows())
		_, err := model.UpdateCourse(courseID, UpdateCourseParams{Title: &title})
		if !errors.Is(err, ErrCourseNotFound) {
			t.Fatalf("error = %v, want %v", err, ErrCourseNotFound)
		}
	})

	t.Run("identical supplied value still updates", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
			WithArgs(title, courseID).
			WillReturnRows(courseSelectRows(Course{
				ID: courseID, OwnerID: "owner000000000000001", Title: title,
				Status: CourseStatusDraft, CreatedAt: now, UpdatedAt: now,
			}))
		_, err := model.UpdateCourse(courseID, UpdateCourseParams{Title: &title})
		if err != nil {
			t.Fatalf("identical value update: %v", err)
		}
	})

	t.Run("publish status transition", func(t *testing.T) {
		model, mock := newCourseModelTest(t)
		status := CourseStatusPublished
		mock.ExpectQuery(`UPDATE "courses".*RETURNING`).
			WithArgs(status, courseID).
			WillReturnRows(courseSelectRows(Course{
				ID: courseID, OwnerID: "owner000000000000001", Title: title,
				Status: CourseStatusPublished, CreatedAt: now, UpdatedAt: now,
			}))
		course, err := model.UpdateCourse(courseID, UpdateCourseParams{Status: &status})
		if err != nil {
			t.Fatalf("publish: %v", err)
		}
		if course.Status != CourseStatusPublished {
			t.Fatalf("status = %s", course.Status)
		}
	})
}
