package models

import (
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func samplePreviewRow(rowNumber int, stem string, answer int) ImportPreviewRow {
	return ImportPreviewRow{
		RowNumber: rowNumber,
		Question:  stem,
		Type:      1,
		Options: map[string]string{
			"1": "Paris",
			"2": "Berlin",
		},
		Answers:             []int{answer},
		OfficialAnswer:      []int{answer},
		AuthoritativeAnswer: []int{answer},
		AnswerReviewStatus:  constants.AnswerReviewUnreviewed,
		RevisionNumber:      1,
	}
}

func TestBuildImportPreviewRowFingerprint_MatchesEquivalentContent(t *testing.T) {
	left := samplePreviewRow(2, "Capital?", 1)
	right := samplePreviewRow(5, "  Capital?  ", 1)

	if BuildImportPreviewRowFingerprint(left) != BuildImportPreviewRowFingerprint(right) {
		t.Fatal("expected equivalent rows to share a fingerprint")
	}
}

func TestApplyImportDuplicateDetection_RejectsLaterFileDuplicate(t *testing.T) {
	rows := []ImportPreviewRow{
		samplePreviewRow(2, "Capital?", 1),
		samplePreviewRow(4, "Capital?", 1),
	}

	kept, errors := ApplyImportDuplicateDetection(rows, nil, ImportFingerprintIndex{})
	if len(kept) != 1 || len(errors) != 1 {
		t.Fatalf("kept=%d errors=%d", len(kept), len(errors))
	}
	if errors[0].Kind != ImportRowErrorKindDuplicate {
		t.Fatalf("kind = %s", errors[0].Kind)
	}
	if errors[0].DuplicateOfRow == nil || *errors[0].DuplicateOfRow != 2 {
		t.Fatalf("duplicate_of_row = %#v", errors[0].DuplicateOfRow)
	}
}

func TestApplyImportDuplicateDetection_RejectsExistingQuizQuestion(t *testing.T) {
	row := samplePreviewRow(3, "Capital?", 1)
	fingerprint := BuildImportPreviewRowFingerprint(row)
	existing := ImportFingerprintIndex{fingerprint: "question-123"}

	kept, errors := ApplyImportDuplicateDetection([]ImportPreviewRow{row}, nil, existing)
	if len(kept) != 0 || len(errors) != 1 {
		t.Fatalf("kept=%d errors=%d", len(kept), len(errors))
	}
	if errors[0].DuplicateQuestionID == nil || *errors[0].DuplicateQuestionID != "question-123" {
		t.Fatalf("duplicate_question_id = %#v", errors[0].DuplicateQuestionID)
	}
}

func TestApplyImportDuplicateDetection_DifferentAnswersAreNotDuplicates(t *testing.T) {
	rows := []ImportPreviewRow{
		samplePreviewRow(2, "Capital?", 1),
		samplePreviewRow(3, "Capital?", 2),
	}

	kept, errors := ApplyImportDuplicateDetection(rows, nil, ImportFingerprintIndex{})
	if len(kept) != 2 || len(errors) != 0 {
		t.Fatalf("kept=%d errors=%d", len(kept), len(errors))
	}
}

func TestFilterQuestionsByQuizDuplicates_RejectsLegacyBatchDuplicate(t *testing.T) {
	questions := []Question{
		{
			Question: "Capital?",
			Type:     1,
			Options: map[string]string{"1": "Paris", "2": "Berlin"},
			Answers: []int{1},
		},
		{
			Question: "Capital?",
			Type:     1,
			Options: map[string]string{"1": "Paris", "2": "Berlin"},
			Answers: []int{1},
		},
	}

	filtered, errors := FilterQuestionsByQuizDuplicates(questions, ImportFingerprintIndex{})
	if len(filtered) != 1 || len(errors) != 1 {
		t.Fatalf("filtered=%d errors=%d", len(filtered), len(errors))
	}
	if errors[0].RowNumber != 3 {
		t.Fatalf("row number = %d", errors[0].RowNumber)
	}
}
