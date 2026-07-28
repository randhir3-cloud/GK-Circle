package utils

import (
	"strings"
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestPreviewQuestionsFromCSV_ValidAndInvalidRows(t *testing.T) {
	rows := []Question{
		{
			Question:      "Capital of France?",
			Type:          "single answer",
			Option1:       "Paris",
			Option2:       "Berlin",
			CorrectAnswer: "1",
		},
		{
			Question:      "",
			Type:          "single answer",
			Option1:       "A",
			Option2:       "B",
			CorrectAnswer: "1",
		},
		{
			Question:      "Only one option",
			Type:          "single answer",
			Option1:       "A",
			CorrectAnswer: "1",
		},
	}

	result, err := PreviewQuestionsFromCSV(rows, "30")
	if err != nil {
		t.Fatalf("PreviewQuestionsFromCSV: %v", err)
	}
	if result.TotalRows != 3 {
		t.Fatalf("TotalRows = %d", result.TotalRows)
	}
	if len(result.ValidRows) != 1 {
		t.Fatalf("ValidRows = %d", len(result.ValidRows))
	}
	if len(result.Errors) != 2 {
		t.Fatalf("Errors = %d", len(result.Errors))
	}

	valid := result.ValidRows[0]
	if valid.RowNumber != 2 {
		t.Fatalf("valid row number = %d", valid.RowNumber)
	}
	if len(valid.Answers) != 1 || valid.Answers[0] != 1 {
		t.Fatalf("answers = %#v", valid.Answers)
	}
	if len(valid.OfficialAnswer) != 1 || valid.OfficialAnswer[0] != 1 {
		t.Fatalf("official_answer = %#v", valid.OfficialAnswer)
	}
	if len(valid.AuthoritativeAnswer) != 1 || valid.AuthoritativeAnswer[0] != 1 {
		t.Fatalf("authoritative_answer = %#v", valid.AuthoritativeAnswer)
	}
	if valid.AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("answer_review_status = %s", valid.AnswerReviewStatus)
	}
	if valid.RevisionNumber != 1 {
		t.Fatalf("revision_number = %d", valid.RevisionNumber)
	}
	if valid.DurationInSeconds != 30 {
		t.Fatalf("duration = %d", valid.DurationInSeconds)
	}
}

func TestPreviewQuestionsFromCSV_EnforcesMaxRows(t *testing.T) {
	rows := make([]Question, constants.MaxRows+1)
	for i := range rows {
		rows[i] = Question{
			Question:      "Q",
			Type:          "single answer",
			Option1:       "A",
			Option2:       "B",
			CorrectAnswer: "1",
		}
	}

	_, err := PreviewQuestionsFromCSV(rows, "30")
	if err == nil {
		t.Fatal("expected max rows error")
	}
	if !strings.Contains(err.Error(), constants.ErrRowsReachesToMaxCount) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractQuestionsFromCSV_EnforcesMaxRows(t *testing.T) {
	rows := make([]Question, constants.MaxRows+1)
	for i := range rows {
		rows[i] = Question{
			Question:      "Q",
			Type:          "single answer",
			Option1:       "A",
			Option2:       "B",
			CorrectAnswer: "1",
		}
	}

	_, err := ExtractQuestionsFromCSV(rows, "30")
	if err == nil {
		t.Fatal("expected max rows error")
	}
	if !strings.Contains(err.Error(), constants.ErrRowsReachesToMaxCount) {
		t.Fatalf("unexpected error: %v", err)
	}
}
