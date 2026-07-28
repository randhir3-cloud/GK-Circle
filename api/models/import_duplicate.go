package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/pkg/structs"
)

const (
	ImportRowErrorKindValidation = "validation"
	ImportRowErrorKindDuplicate  = "duplicate"
)

// ImportQuestionFingerprint is a deterministic content key for duplicate detection.
type ImportQuestionFingerprint string

// ImportFingerprintIndex maps a question fingerprint to an existing question ID.
type ImportFingerprintIndex map[ImportQuestionFingerprint]string

// BuildImportPreviewRowFingerprint fingerprints stem, type, options, and operational answers.
func BuildImportPreviewRowFingerprint(row ImportPreviewRow) ImportQuestionFingerprint {
	return buildImportQuestionFingerprint(row.Type, row.Question, row.Options, row.Answers)
}

// BuildQuestionFingerprint fingerprints a bank question for duplicate comparison.
func BuildQuestionFingerprint(question Question) ImportQuestionFingerprint {
	return buildImportQuestionFingerprint(question.Type, question.Question, question.Options, question.Answers)
}

// BuildQuestionAnalyticsFingerprint fingerprints an existing quiz question row.
func BuildQuestionAnalyticsFingerprint(question structs.QuestionAnalytics) (ImportQuestionFingerprint, error) {
	options := question.Options
	if len(options) == 0 && len(question.RawOptions) > 0 {
		if err := json.Unmarshal(question.RawOptions, &options); err != nil {
			return "", err
		}
	}

	answers, err := parseStoredAnswerKeys(question.CorrectAnswer)
	if err != nil {
		return "", err
	}

	return buildImportQuestionFingerprint(question.QuestionTypeID, question.Question, options, answers), nil
}

func buildImportQuestionFingerprint(
	questionType int,
	stem string,
	options map[string]string,
	answers []int,
) ImportQuestionFingerprint {
	return ImportQuestionFingerprint(fmt.Sprintf(
		"t:%d|q:%s|o:%s|a:%s",
		questionType,
		normalizeImportStem(stem),
		canonicalImportOptions(options),
		canonicalImportAnswers(answers),
	))
}

func normalizeImportStem(stem string) string {
	return strings.TrimSpace(stem)
}

func canonicalImportOptions(options map[string]string) string {
	if len(options) == 0 {
		return ""
	}

	keys := make([]int, 0, len(options))
	for key := range options {
		parsed, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil {
			continue
		}
		keys = append(keys, parsed)
	}
	sort.Ints(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(options[strconv.Itoa(key)])
		parts = append(parts, fmt.Sprintf("%d=%s", key, value))
	}
	return strings.Join(parts, "|")
}

func canonicalImportAnswers(answers []int) string {
	if len(answers) == 0 {
		return ""
	}

	sorted := append([]int(nil), answers...)
	sort.Ints(sorted)
	parts := make([]string, 0, len(sorted))
	for _, answer := range sorted {
		parts = append(parts, strconv.Itoa(answer))
	}
	return strings.Join(parts, ",")
}

func parseStoredAnswerKeys(raw string) ([]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var answers []int
	if err := json.Unmarshal([]byte(raw), &answers); err == nil {
		return answers, nil
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return []int{parsed}, nil
}

// ApplyImportDuplicateDetection moves duplicate preview rows into structured row errors.
// The first occurrence in the CSV file wins; later file duplicates and quiz duplicates are rejected.
func ApplyImportDuplicateDetection(
	validRows []ImportPreviewRow,
	errors []ImportRowError,
	existing ImportFingerprintIndex,
) ([]ImportPreviewRow, []ImportRowError) {
	if len(validRows) == 0 {
		return validRows, errors
	}

	seenInFile := make(map[ImportQuestionFingerprint]int, len(validRows))
	kept := make([]ImportPreviewRow, 0, len(validRows))

	for _, row := range validRows {
		fingerprint := BuildImportPreviewRowFingerprint(row)

		if firstRow, ok := seenInFile[fingerprint]; ok {
			first := firstRow
			errors = append(errors, ImportRowError{
				RowNumber:       row.RowNumber,
				Messages:        []string{fmt.Sprintf(constants.ErrImportDuplicateInFile, firstRow)},
				Kind:            ImportRowErrorKindDuplicate,
				DuplicateOfRow:  &first,
			})
			continue
		}

		if questionID, ok := existing[fingerprint]; ok {
			id := questionID
			errors = append(errors, ImportRowError{
				RowNumber:           row.RowNumber,
				Messages:              []string{fmt.Sprintf(constants.ErrImportDuplicateInQuiz, questionID)},
				Kind:                  ImportRowErrorKindDuplicate,
				DuplicateQuestionID: &id,
			})
			continue
		}

		seenInFile[fingerprint] = row.RowNumber
		kept = append(kept, row)
	}

	return kept, errors
}

// FindImportDuplicateErrors checks preview rows against the current quiz fingerprint index.
func FindImportDuplicateErrors(
	rows []ImportPreviewRow,
	existing ImportFingerprintIndex,
) []ImportRowError {
	_, errors := ApplyImportDuplicateDetection(rows, nil, existing)
	return errors
}

// FilterQuestionsByQuizDuplicates removes duplicate questions before legacy import append.
func FilterQuestionsByQuizDuplicates(
	questions []Question,
	existing ImportFingerprintIndex,
) ([]Question, []ImportRowError) {
	if len(questions) == 0 {
		return questions, nil
	}

	seenInFile := make(map[ImportQuestionFingerprint]int, len(questions))
	kept := make([]Question, 0, len(questions))
	var errors []ImportRowError

	for index, question := range questions {
		rowNumber := index + 2
		fingerprint := BuildQuestionFingerprint(question)

		if firstRow, ok := seenInFile[fingerprint]; ok {
			first := firstRow
			errors = append(errors, ImportRowError{
				RowNumber:      rowNumber,
				Messages:       []string{fmt.Sprintf(constants.ErrImportDuplicateInFile, firstRow)},
				Kind:           ImportRowErrorKindDuplicate,
				DuplicateOfRow: &first,
			})
			continue
		}

		if questionID, ok := existing[fingerprint]; ok {
			id := questionID
			errors = append(errors, ImportRowError{
				RowNumber:           rowNumber,
				Messages:              []string{fmt.Sprintf(constants.ErrImportDuplicateInQuiz, questionID)},
				Kind:                  ImportRowErrorKindDuplicate,
				DuplicateQuestionID: &id,
			})
			continue
		}

		seenInFile[fingerprint] = rowNumber
		kept = append(kept, question)
	}

	return kept, errors
}
