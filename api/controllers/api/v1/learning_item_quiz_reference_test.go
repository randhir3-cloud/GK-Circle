package v1

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

func TestValidateCreateAdminLearningItemQuizReference(t *testing.T) {
	courseID := uuid.MustParse("019c01c6-b8b7-7f4a-9e7a-f62650d5e481")
	nodeID := uuid.MustParse("019c01c8-4bdd-78e2-a366-690bfd600280")
	quizID := uuid.MustParse("019c02a0-1111-7000-8000-000000000101")

	t.Run("requires quiz ID for quiz reference", func(t *testing.T) {
		_, err := validateCreateAdminLearningItem(
			courseID,
			nodeID,
			"admin-1",
			structs.ReqCreateAdminLearningItem{
				Title:    "Polity test",
				ItemType: string(models.LearningItemTypeQuizRef),
			},
		)
		if !errors.Is(err, models.ErrLearningItemQuizRequired) {
			t.Fatalf("error = %v, want %v", err, models.ErrLearningItemQuizRequired)
		}
	})

	t.Run("accepts validated quiz ID", func(t *testing.T) {
		params, err := validateCreateAdminLearningItem(
			courseID,
			nodeID,
			"admin-1",
			structs.ReqCreateAdminLearningItem{
				Title:    "Polity test",
				ItemType: string(models.LearningItemTypeQuizRef),
				QuizID: structs.OptionalString{
					Present: true,
					Value:   quizID.String(),
				},
			},
		)
		if err != nil {
			t.Fatalf("validate: %v", err)
		}
		if params.QuizID == nil || *params.QuizID != quizID {
			t.Fatalf("quiz ID = %v, want %s", params.QuizID, quizID)
		}
		if params.ActorID != "admin-1" {
			t.Fatalf("actor ID = %q, want admin-1", params.ActorID)
		}
	})

	t.Run("rejects quiz ID on non-quiz content", func(t *testing.T) {
		_, err := validateCreateAdminLearningItem(
			courseID,
			nodeID,
			"admin-1",
			structs.ReqCreateAdminLearningItem{
				Title:    "Article",
				ItemType: string(models.LearningItemTypeArticle),
				QuizID: structs.OptionalString{
					Present: true,
					Value:   quizID.String(),
				},
			},
		)
		if !errors.Is(err, models.ErrLearningItemQuizForbidden) {
			t.Fatalf("error = %v, want %v", err, models.ErrLearningItemQuizForbidden)
		}
	})
}

func TestValidateUpdateAdminLearningItemQuizReference(t *testing.T) {
	quizID := uuid.MustParse("019c02a0-1111-7000-8000-000000000101")
	quizType := structs.OptionalString{
		Present: true,
		Value:   string(models.LearningItemTypeQuizRef),
	}

	_, err := validateUpdateAdminLearningItem(
		"admin-1",
		structs.ReqUpdateAdminLearningItem{ItemType: quizType},
	)
	if !errors.Is(err, models.ErrLearningItemQuizRequired) {
		t.Fatalf("missing quiz error = %v, want %v", err, models.ErrLearningItemQuizRequired)
	}

	params, err := validateUpdateAdminLearningItem(
		"admin-1",
		structs.ReqUpdateAdminLearningItem{
			ItemType: quizType,
			QuizID: structs.OptionalString{
				Present: true,
				Value:   quizID.String(),
			},
		},
	)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !params.QuizID.Present || params.QuizID.Null || params.QuizID.Value != quizID {
		t.Fatalf("quiz update = %+v", params.QuizID)
	}
}
