package models

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
)

func TestHasQuizReviewPreviewAccess(t *testing.T) {
	cfg := config.AppConfig{}
	cfg.Quiz.PublicQuizAdminEmails = []string{"admin@gkcircle.com"}

	sharedDB, sharedMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sharedDB.Close() })
	sharedModel := InitSharedQuizzesModel(goqu.New("postgres", sharedDB), nil)

	accessCtx := PlayedQuizReviewAccessContext{
		UserID:         "participant-1",
		SessionAdminID: "host-1",
		QuizID:         uuid.NewString(),
	}

	if allowed, err := HasQuizReviewPreviewAccess("participant-1", nil, &cfg, sharedModel, accessCtx); err != nil || !allowed {
		t.Fatalf("participant access = %v, err = %v", allowed, err)
	}

	if allowed, err := HasQuizReviewPreviewAccess("host-1", nil, &cfg, sharedModel, accessCtx); err != nil || !allowed {
		t.Fatalf("host access = %v, err = %v", allowed, err)
	}

	if allowed, err := HasQuizReviewPreviewAccess("stranger", nil, &cfg, sharedModel, accessCtx); err != nil || allowed {
		t.Fatalf("stranger without editor = %v, err = %v", allowed, err)
	}

	editor := &User{ID: "editor-1", Email: "editor@example.com"}
	sharedMock.ExpectQuery(`SELECT (.+) FROM "quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))
	sharedMock.ExpectQuery(`SELECT (.+) FROM "shared_quizzes"`).
		WillReturnRows(sqlmock.NewRows([]string{"permission"}).AddRow("read"))

	if allowed, err := HasQuizReviewPreviewAccess("editor-1", editor, &cfg, sharedModel, accessCtx); err != nil || !allowed {
		t.Fatalf("shared read editor access = %v, err = %v", allowed, err)
	}
}

func TestUserPlayedQuizModel_CanAccessReview_NotFound(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	model := InitUserPlayedQuizModel(goqu.New("postgres", sqlDB))
	playedQuizID := uuid.NewString()

	mock.ExpectQuery(`SELECT (.+) FROM "user_played_quizzes"`).
		WillReturnError(sql.ErrNoRows)

	allowed, err := model.CanAccessReview("user-1", nil, &config.AppConfig{}, InitSharedQuizzesModel(goqu.New("postgres", sqlDB), nil), playedQuizID)
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v (allowed=%v)", err, allowed)
	}
}
