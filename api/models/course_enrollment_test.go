package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
)

func newCourseEnrollmentModel(t *testing.T) (*CourseEnrollmentModel, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return InitCourseEnrollmentModel(goqu.New("postgres", sqlDB)), mock
}

func TestCourseEnrollmentRequireUserEnrolled(t *testing.T) {
	model, mock := newCourseEnrollmentModel(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	userID := "owner000000000000099"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "course_enrollments" WHERE`)).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	err := model.RequireUserEnrolled(userID, courseID)
	if err != ErrCourseEnrollmentRequired {
		t.Fatalf("err = %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "course_enrollments" WHERE`)).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	if err := model.RequireUserEnrolled(userID, courseID); err != nil {
		t.Fatalf("enrolled err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestCourseEnrollmentEnrollUserIdempotent(t *testing.T) {
	model, mock := newCourseEnrollmentModel(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	userID := "owner000000000000099"
	enrollmentID := uuid.MustParse("019c0e00-1111-7000-8000-000000000001")
	now := time.Date(2026, time.July, 27, 1, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "status" FROM "courses" WHERE`)).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PUBLISHED"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id" FROM "course_enrollments" WHERE`)).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(enrollmentID))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id", "course_id", "user_id", "enrolled_at" FROM "course_enrollments" WHERE`)).
		WithArgs(courseID, userID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "course_id", "user_id", "enrolled_at"}).
			AddRow(enrollmentID, courseID, userID, now))

	got, err := model.EnrollUser(userID, courseID)
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if got.ID != enrollmentID || got.UserID != userID {
		t.Fatalf("got = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestCourseEnrollmentEnrollUserDraftRejected(t *testing.T) {
	model, mock := newCourseEnrollmentModel(t)
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e482")
	userID := "owner000000000000099"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "status" FROM "courses" WHERE`)).
		WithArgs(courseID, uint(1)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("DRAFT"))

	_, err := model.EnrollUser(userID, courseID)
	if err != ErrCourseNotPublished {
		t.Fatalf("err = %v", err)
	}
}
