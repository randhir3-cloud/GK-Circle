package utils

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	quizUtilsHelper "github.com/randhir3-cloud/GK-Circle-v2/api/helpers/utils"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/google/uuid"
)

// CSVImportPreviewResult is the structured validation preview for EXAM-P3-T01.
type CSVImportPreviewResult struct {
	ValidRows []models.ImportPreviewRow
	Errors    []models.ImportRowError
	TotalRows int
}

// PreviewQuestionsFromCSV validates each CSV row independently and returns a
// preview payload. Unlike ExtractQuestionsFromCSV it does not fail the entire
// file when some rows are invalid, and it never persists questions.
func PreviewQuestionsFromCSV(questions []Question, questionTimeLimit string) (CSVImportPreviewResult, error) {
	result := CSVImportPreviewResult{}

	if len(questions) == 0 {
		return result, fmt.Errorf(constants.ErrEmptyFile)
	}
	if len(questions) > constants.MaxRows {
		return result, fmt.Errorf("%s (max %d)", constants.ErrRowsReachesToMaxCount, constants.MaxRows)
	}

	duration, err := strconv.Atoi(strings.TrimSpace(questionTimeLimit))
	if err != nil || duration <= 0 {
		return result, fmt.Errorf(constants.ErrInvalidQuestionTimeLimit)
	}

	result.TotalRows = len(questions)

	for i, u := range questions {
		rowNo := i + 2
		question, rowIssues := validateCSVQuestionRow(u, duration, i+1)
		if len(rowIssues) > 0 {
			result.Errors = append(result.Errors, models.ImportRowError{
				RowNumber: rowNo,
				Messages:  rowIssues,
			})
			continue
		}

		previewRow, err := models.BuildImportPreviewRow(rowNo, question)
		if err != nil {
			result.Errors = append(result.Errors, models.ImportRowError{
				RowNumber: rowNo,
				Messages:  []string{err.Error()},
			})
			continue
		}
		result.ValidRows = append(result.ValidRows, previewRow)
	}

	return result, nil
}

func validateCSVQuestionRow(u Question, duration int, orderNumber int) (models.Question, []string) {
	var rowIssues []string

	if strings.TrimSpace(u.Question) == "" {
		rowIssues = append(rowIssues, constants.ErrEmptyQuestionText)
	}

	questionType, typeErr := quizUtilsHelper.CheckQuestionType(strings.TrimSpace(u.Type))
	if typeErr != nil {
		rowIssues = append(rowIssues, fmt.Sprintf("%s (got %q, allowed: %s, %s)", constants.ErrQuestionType, u.Type, constants.SingleAnswerString, constants.SurveyString))
	}

	options := make(map[string]string)
	for idx, opt := range []string{u.Option1, u.Option2, u.Option3, u.Option4, u.Option5} {
		if strings.TrimSpace(opt) != "" {
			options[strconv.Itoa(idx+1)] = opt
		}
	}
	if len(options) < 2 {
		rowIssues = append(rowIssues, constants.ErrInsufficientOptions)
	}

	answers := []int{}
	correctRaw := strings.TrimSpace(u.CorrectAnswer)
	if correctRaw == "" {
		rowIssues = append(rowIssues, constants.ErrEmptyCorrectAnswer)
	} else {
		for _, a := range strings.Split(correctRaw, "|") {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			answerInt, convErr := strconv.Atoi(a)
			if convErr != nil {
				rowIssues = append(rowIssues, fmt.Sprintf("%s (got %q)", constants.ErrInvalidCorrectAnswer, a))
				continue
			}
			if _, ok := options[strconv.Itoa(answerInt)]; !ok {
				rowIssues = append(rowIssues, fmt.Sprintf("%s (option %d does not exist)", constants.ErrInvalidCorrectAnswer, answerInt))
				continue
			}
			answers = append(answers, answerInt)
		}
	}

	if typeErr == nil {
		switch questionType {
		case constants.SingleAnswer:
			if len(answers) != 1 {
				rowIssues = append(rowIssues, constants.ErrSingleAnswerLength)
			}
		case constants.Survey:
			if len(answers) < 1 {
				rowIssues = append(rowIssues, constants.ErrSurveyAnswerLength)
			}
		}
	}

	questionMedia, questionMediaOK := normalizeMedia(u.QuestionMedia)
	if !questionMediaOK {
		rowIssues = append(rowIssues, fmt.Sprintf("%s (got %q)", constants.ErrInvalidQuestionMedia, u.QuestionMedia))
	}
	optionsMedia, optionsMediaOK := normalizeMedia(u.OptionsMedia)
	if !optionsMediaOK {
		rowIssues = append(rowIssues, fmt.Sprintf("%s (got %q)", constants.ErrInvalidOptionsMedia, u.OptionsMedia))
	}

	points := 1
	if strings.TrimSpace(u.Points) != "" {
		parsedPoints, convErr := strconv.Atoi(strings.TrimSpace(u.Points))
		if convErr != nil || parsedPoints <= 0 {
			rowIssues = append(rowIssues, fmt.Sprintf("%s (got %q)", constants.ErrInvalidPoints, u.Points))
		} else {
			points = parsedPoints
		}
	}

	if len(rowIssues) > 0 {
		return models.Question{}, rowIssues
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return models.Question{}, []string{err.Error()}
	}

	return models.Question{
		ID:                id,
		Question:          u.Question,
		Type:              questionType,
		Options:           options,
		Answers:           answers,
		Points:            int16(points),
		DurationInSeconds: duration,
		OrderNumber:       orderNumber,
		QuestionMedia:     questionMedia,
		OptionsMedia:      optionsMedia,
		Resource:          sql.NullString{String: u.Resource, Valid: true},
	}, nil
}
