package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

const CoursesTable = "courses"

type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "DRAFT"
	CourseStatusPublished CourseStatus = "PUBLISHED"
	CourseStatusArchived  CourseStatus = "ARCHIVED"
)

var (
	ErrCourseOwnerRequired  = errors.New("course owner is required")
	ErrCourseTitleRequired  = errors.New("course title is required")
	ErrCourseNotFound       = errors.New("course not found")
	ErrCourseUpdateRequired = errors.New("course update requires at least one field")
	ErrCourseStatusInvalid  = errors.New("course status is invalid")
	ErrCourseNotPublished   = errors.New("course is not published")
)

var courseColumns = []interface{}{
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
}

// Course is the root educational entity. Curriculum hierarchy, publishing
// validation, enrollment, branding, pricing, and analytics are intentionally
// outside COURSE-P1-T02.
type Course struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	OwnerID          string         `json:"owner_id" db:"owner_id"`
	Title            string         `json:"title" db:"title"`
	ShortDescription sql.NullString `json:"short_description,omitempty" db:"short_description"`
	Language         sql.NullString `json:"language,omitempty" db:"language"`
	Difficulty       sql.NullString `json:"difficulty,omitempty" db:"difficulty"`
	Visibility       sql.NullString `json:"visibility,omitempty" db:"visibility"`
	Status           CourseStatus   `json:"status" db:"status"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateCourseParams struct {
	OwnerID          string
	Title            string
	ShortDescription string
	Language         string
	Difficulty       string
	Visibility       string
}

// OptionalNullableString carries presence for partial Course updates.
// Present=false means omitted; Present+Null means set SQL NULL; Present with
// a non-null value applies the trimmed empty-string-to-NULL rule.
type OptionalNullableString struct {
	Present bool
	Null    bool
	Value   string
}

type UpdateCourseParams struct {
	Title            *string
	ShortDescription OptionalNullableString
	Language         OptionalNullableString
	Difficulty       OptionalNullableString
	Visibility       OptionalNullableString
	Status           *CourseStatus
}

// CourseModel implements Course root persistence using the repository's
// established goqu-backed model convention.
type CourseModel struct {
	db *goqu.Database
}

func InitCourseModel(goquDB *goqu.Database) *CourseModel {
	return &CourseModel{db: goquDB}
}

func (model *CourseModel) CreateCourse(params CreateCourseParams) (Course, error) {
	var course Course

	ownerID := strings.TrimSpace(params.OwnerID)
	if ownerID == "" {
		return course, ErrCourseOwnerRequired
	}

	title := strings.TrimSpace(params.Title)
	if title == "" {
		return course, ErrCourseTitleRequired
	}

	courseID, err := uuid.NewUUID()
	if err != nil {
		return course, err
	}

	found, err := model.db.Insert(CoursesTable).
		Rows(goqu.Record{
			"id":                courseID,
			"owner_id":          ownerID,
			"title":             title,
			"short_description": nullableTrimmedString(params.ShortDescription),
			"language":          nullableTrimmedString(params.Language),
			"difficulty":        nullableTrimmedString(params.Difficulty),
			"visibility":        nullableTrimmedString(params.Visibility),
			"status":            CourseStatusDraft,
		}).
		Returning(courseColumns...).
		Prepared(true).
		Executor().
		ScanStruct(&course)
	if err != nil {
		return course, err
	}
	if !found {
		return course, sql.ErrNoRows
	}

	return course, nil
}

func (model *CourseModel) GetCourseByID(courseID uuid.UUID) (Course, error) {
	var course Course

	found, err := model.db.From(CoursesTable).
		Select(courseColumns...).
		Where(goqu.Ex{"id": courseID}).
		Limit(1).
		Prepared(true).
		ScanStruct(&course)
	if err != nil {
		return course, err
	}
	if !found {
		return course, ErrCourseNotFound
	}

	return course, nil
}

func (model *CourseModel) ListCourses() ([]Course, error) {
	courses := make([]Course, 0)
	err := model.db.From(CoursesTable).
		Select(courseColumns...).
		Order(goqu.I("created_at").Desc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// ListPublishedCourses returns courses with status PUBLISHED, newest first.
func (model *CourseModel) ListPublishedCourses() ([]Course, error) {
	courses := make([]Course, 0)
	err := model.db.From(CoursesTable).
		Select(courseColumns...).
		Where(goqu.Ex{"status": CourseStatusPublished}).
		Order(goqu.I("created_at").Desc(), goqu.I("id").Asc()).
		Prepared(true).
		ScanStructs(&courses)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// RequirePublishedCourse returns ErrCourseNotFound or ErrCourseNotPublished.
func (model *CourseModel) RequirePublishedCourse(courseID uuid.UUID) (Course, error) {
	course, err := model.GetCourseByID(courseID)
	if err != nil {
		return course, err
	}
	if course.Status != CourseStatusPublished {
		return course, ErrCourseNotPublished
	}
	return course, nil
}

func (model *CourseModel) UpdateCourse(courseID uuid.UUID, params UpdateCourseParams) (Course, error) {
	var course Course

	record := goqu.Record{}
	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return course, ErrCourseTitleRequired
		}
		record["title"] = title
	}
	if params.ShortDescription.Present {
		record["short_description"] = optionalNullableToSQL(params.ShortDescription)
	}
	if params.Language.Present {
		record["language"] = optionalNullableToSQL(params.Language)
	}
	if params.Difficulty.Present {
		record["difficulty"] = optionalNullableToSQL(params.Difficulty)
	}
	if params.Visibility.Present {
		record["visibility"] = optionalNullableToSQL(params.Visibility)
	}
	if params.Status != nil {
		status := *params.Status
		switch status {
		case CourseStatusDraft, CourseStatusPublished, CourseStatusArchived:
			record["status"] = status
		default:
			return course, ErrCourseStatusInvalid
		}
	}
	if len(record) == 0 {
		return course, ErrCourseUpdateRequired
	}
	record["updated_at"] = goqu.L("now()")

	found, err := model.db.Update(CoursesTable).
		Set(record).
		Where(goqu.Ex{"id": courseID}).
		Returning(courseColumns...).
		Prepared(true).
		Executor().
		ScanStruct(&course)
	if err != nil {
		return course, err
	}
	if !found {
		return course, ErrCourseNotFound
	}
	return course, nil
}

func optionalNullableToSQL(value OptionalNullableString) interface{} {
	if value.Null {
		return nil
	}
	trimmed := strings.TrimSpace(value.Value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableTrimmedString(value string) sql.NullString {
	trimmed := strings.TrimSpace(value)
	return sql.NullString{String: trimmed, Valid: trimmed != ""}
}
