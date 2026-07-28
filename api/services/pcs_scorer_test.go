package services

import (
	"testing"

	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestScorePCSAttempt_CorrectIncorrectUnansweredNegativeMarks(t *testing.T) {
	q1 := uuid.New()
	q2 := uuid.New()
	q3 := uuid.New()
	points := int16(2)

	result := ScorePCSAttempt([]PCSQuestionScoreInput{
		{
			QuestionID: q1, Type: constants.SingleAnswer, Points: &points,
			SelectedOptions: []int{1}, AuthoritativeAnswer: []int{1},
		},
		{
			QuestionID: q2, Type: constants.SingleAnswer, Points: &points,
			SelectedOptions: []int{2}, AuthoritativeAnswer: []int{1},
		},
		{
			QuestionID: q3, Type: constants.SingleAnswer, Points: &points,
			SelectedOptions: []int{}, AuthoritativeAnswer: []int{1},
		},
	}, PCSScoreConfig{NegativeMarksPerQuestion: 0.5})

	if result.MaxScore != 6 {
		t.Fatalf("max = %v", result.MaxScore)
	}
	// 2 - 0.5 + 0 = 1.5
	if result.TotalScore != 1.5 {
		t.Fatalf("total = %v", result.TotalScore)
	}
	if result.CorrectCount != 1 || result.IncorrectCount != 1 || result.UnansweredCount != 1 {
		t.Fatalf("counts = %+v", result)
	}
}

func TestScorePCSAttempt_PrefersAuthoritativeOverOperational(t *testing.T) {
	qid := uuid.New()
	points := int16(1)
	result := ScorePCSAttempt([]PCSQuestionScoreInput{
		{
			QuestionID: qid, Type: constants.SingleAnswer, Points: &points,
			SelectedOptions:     []int{2},
			AuthoritativeAnswer: []int{2},
			OperationalAnswers:  []int{1}, // live/operational differs; must not win
		},
	}, PCSScoreConfig{})
	if result.CorrectCount != 1 {
		t.Fatalf("expected authoritative key to win: %+v", result.QuestionResults[0])
	}
	if result.QuestionResults[0].KeySource != ScoringKeyAuthoritative {
		t.Fatalf("key source = %s", result.QuestionResults[0].KeySource)
	}
}

func TestScorePCSAttempt_SurveyUnscored(t *testing.T) {
	qid := uuid.New()
	points := int16(5)
	result := ScorePCSAttempt([]PCSQuestionScoreInput{
		{
			QuestionID: qid, Type: constants.Survey, Points: &points,
			SelectedOptions: []int{1, 2}, AuthoritativeAnswer: []int{1, 2},
		},
	}, PCSScoreConfig{NegativeMarksPerQuestion: 1})
	if result.MaxScore != 0 || result.TotalScore != 0 || result.UnscoredCount != 1 {
		t.Fatalf("survey scored unexpectedly: %+v", result)
	}
	if result.QuestionResults[0].IsCorrect != nil {
		t.Fatal("survey should leave is_correct null")
	}
}

func TestScorePCSAttempt_FloorAtZero(t *testing.T) {
	qid := uuid.New()
	points := int16(1)
	result := ScorePCSAttempt([]PCSQuestionScoreInput{
		{
			QuestionID: qid, Type: constants.SingleAnswer, Points: &points,
			SelectedOptions: []int{2}, AuthoritativeAnswer: []int{1},
		},
	}, PCSScoreConfig{NegativeMarksPerQuestion: 2})
	if result.TotalScore != 0 {
		t.Fatalf("expected floor 0, got %v", result.TotalScore)
	}
}

func TestScorePCSAttempt_FallbackOperationalKey(t *testing.T) {
	qid := uuid.New()
	result := ScorePCSAttempt([]PCSQuestionScoreInput{
		{
			QuestionID: qid, Type: constants.SingleAnswer,
			SelectedOptions:    []int{1},
			OperationalAnswers: []int{1},
		},
	}, PCSScoreConfig{})
	if result.QuestionResults[0].KeySource != ScoringKeyOperational {
		t.Fatalf("source = %s", result.QuestionResults[0].KeySource)
	}
	if result.CorrectCount != 1 {
		t.Fatal("expected correct")
	}
}
