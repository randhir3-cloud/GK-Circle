package utils

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/jszwec/csvutil"
)

type Question struct {
	Question      string `csv:"Question Text"`
	Type          string `csv:"Question Type"`
	Points        string `csv:"Points,omitempty"`
	Option1       string `csv:"Option 1"`
	Option2       string `csv:"Option 2"`
	Option3       string `csv:"Option 3"`
	Option4       string `csv:"Option 4"`
	Option5       string `csv:"Option 5"`
	CorrectAnswer string `csv:"Correct Answer"`
	QuestionMedia string `csv:"Question Media"`
	OptionsMedia  string `csv:"Options Media"`
	Resource      string `csv:"Resource"`
}

func ValidateCSVFileFormat(fileName string) ([]Question, error) {
	var questions []Question

	// Open the CSV file
	file, err := os.Open(fileName)
	if err != nil {
		return questions, err
	}
	defer file.Close()

	// Create a new CSV reader
	csvData, err := io.ReadAll(file)
	if err != nil {
		return questions, err
	}

	if err := csvutil.Unmarshal(csvData, &questions); err != nil {
		return questions, err
	}

	if len(questions) == 0 {
		return questions, fmt.Errorf(constants.ErrEmptyFile)
	}

	return questions, nil
}

// normalizeMedia trims and lowercases a media value and reports whether it is
// allowed. An empty value defaults to "text" to match manual question creation.
func normalizeMedia(media string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(media))
	if normalized == "" {
		return constants.MediaText, true
	}
	switch normalized {
	case constants.MediaText, constants.MediaImage, constants.MediaCode:
		return normalized, true
	default:
		return normalized, false
	}
}

func ExtractQuestionsFromCSV(questions []Question, questionTimeLimit string) ([]models.Question, error) {
	// Duration comes solely from configuration; reject if it is missing or invalid.
	duration, err := strconv.Atoi(strings.TrimSpace(questionTimeLimit))
	if err != nil || duration <= 0 {
		return nil, fmt.Errorf(constants.ErrInvalidQuestionTimeLimit)
	}

	if len(questions) > constants.MaxRows {
		return nil, fmt.Errorf("%s (max %d)", constants.ErrRowsReachesToMaxCount, constants.MaxRows)
	}

	var validQuestions []models.Question
	var rowErrors []string

	for i, u := range questions {
		// Row number as seen by the user in a spreadsheet (header is row 1).
		rowNo := i + 2

		question, rowIssues := validateCSVQuestionRow(u, duration, i+1)
		if len(rowIssues) > 0 {
			rowErrors = append(rowErrors, fmt.Sprintf("row %d: %s", rowNo, strings.Join(rowIssues, "; ")))
			continue
		}

		validQuestions = append(validQuestions, question)
	}

	if len(rowErrors) > 0 {
		return nil, fmt.Errorf("%s %s", constants.ErrInvalidCSVRows, strings.Join(rowErrors, " | "))
	}

	return validQuestions, nil
}
