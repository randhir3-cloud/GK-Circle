package models

import (
	"errors"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

const CourseEnrollmentsTable = "course_enrollments"

var (
	ErrCourseEnrollmentRequired     = errors.New("course enrollment required")
	ErrCourseEnrollmentUserRequired = errors.New("course enrollment user is required")
)

// CourseEnrollment is the persisted learner Course access grant.
type CourseEnrollment struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CourseID   uuid.UUID `json:"course_id" db:"course_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	EnrolledAt time.Time `json:"enrolled_at" db:"enrolled_at"`
}

type CourseEnrollmentModel struct {
	db *goqu.Database
}

func InitCourseEnrollmentModel(db *goqu.Database) *CourseEnrollmentModel {
	return &CourseEnrollmentModel{db: db}
}

// IsUserEnrolled reports whether userID has an enrollment row for courseID.
func (model *CourseEnrollmentModel) IsUserEnrolled(userID string, courseID uuid.UUID) (bool, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return false, ErrCourseEnrollmentUserRequired
	}

	found, err := model.db.From(CourseEnrollmentsTable).
		Select(goqu.C("id")).
		Where(goqu.Ex{
			"course_id": courseID,
			"user_id":   userID,
		}).
		Limit(1).
		Prepared(true).
		ScanVal(new(uuid.UUID))
	if err != nil {
		return false, err
	}
	return found, nil
}

// RequireUserEnrolled returns ErrCourseEnrollmentRequired when the user is not enrolled.
func (model *CourseEnrollmentModel) RequireUserEnrolled(userID string, courseID uuid.UUID) error {
	enrolled, err := model.IsUserEnrolled(userID, courseID)
	if err != nil {
		return err
	}
	if !enrolled {
		return ErrCourseEnrollmentRequired
	}
	return nil
}

// EnrollUser creates an enrollment idempotently. The Course must exist and be PUBLISHED.
func (model *CourseEnrollmentModel) EnrollUser(userID string, courseID uuid.UUID) (CourseEnrollment, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return CourseEnrollment{}, ErrCourseEnrollmentUserRequired
	}

	var status string
	exists, err := model.db.From(CoursesTable).
		Select(goqu.C("status")).
		Where(goqu.Ex{"id": courseID}).
		Limit(1).
		Prepared(true).
		ScanVal(&status)
	if err != nil {
		return CourseEnrollment{}, err
	}
	if !exists {
		return CourseEnrollment{}, ErrCourseNotFound
	}
	if CourseStatus(status) != CourseStatusPublished {
		return CourseEnrollment{}, ErrCourseNotPublished
	}

	enrolled, err := model.IsUserEnrolled(userID, courseID)
	if err != nil {
		return CourseEnrollment{}, err
	}
	if enrolled {
		var existing CourseEnrollment
		found, err := model.db.From(CourseEnrollmentsTable).
			Select("id", "course_id", "user_id", "enrolled_at").
			Where(goqu.Ex{
				"course_id": courseID,
				"user_id":   userID,
			}).
			Limit(1).
			Prepared(true).
			ScanStruct(&existing)
		if err != nil {
			return CourseEnrollment{}, err
		}
		if !found {
			return CourseEnrollment{}, ErrCourseEnrollmentRequired
		}
		return existing, nil
	}

	enrollment := CourseEnrollment{
		ID:         uuid.New(),
		CourseID:   courseID,
		UserID:     userID,
		EnrolledAt: time.Now().UTC(),
	}
	_, err = model.db.Insert(CourseEnrollmentsTable).Rows(goqu.Record{
		"id":          enrollment.ID,
		"course_id":   enrollment.CourseID,
		"user_id":     enrollment.UserID,
		"enrolled_at": enrollment.EnrolledAt,
	}).Prepared(true).Executor().Exec()
	if err != nil {
		return CourseEnrollment{}, err
	}
	return enrollment, nil
}

// UnenrollUser removes an enrollment if present. Missing enrollment is a noop.
func (model *CourseEnrollmentModel) UnenrollUser(userID string, courseID uuid.UUID) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrCourseEnrollmentUserRequired
	}
	_, err := model.db.Delete(CourseEnrollmentsTable).Where(goqu.Ex{
		"course_id": courseID,
		"user_id":   userID,
	}).Prepared(true).Executor().Exec()
	return err
}
