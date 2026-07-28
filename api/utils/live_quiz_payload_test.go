package utils

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
	"github.com/randhir3-cloud/GK-Circle-v2/api/security"
)

func TestBuildLiveQuestionDeliveryPayload_ExcludesAnswerKeys(t *testing.T) {
	question := models.Question{
		ID:                uuid.New(),
		QuizId:            uuid.New(),
		OrderNumber:       1,
		DurationInSeconds: 30,
		Question:          "Capital?",
		Options:           map[string]string{"1": "Paris"},
		Answers:           []int{1},
		QuestionMedia:     "text",
		OptionsMedia:      "text",
		Resource:          EmptySQLNullString(),
	}

	payload := BuildLiveQuestionDeliveryPayload(question, question.CreatedAt, 5, 12)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := security.AssertNoSensitiveAnswerKeyFields(body); err != nil {
		t.Fatalf("live delivery payload leaked answer keys: %v", err)
	}
}

func TestBuildLiveScoreboardPayload_IncludesAnswersAfterQuestion(t *testing.T) {
	question := models.Question{
		ID:            uuid.New(),
		QuizId:        uuid.New(),
		OrderNumber:   1,
		Question:      "Capital?",
		Options:       map[string]string{"1": "Paris"},
		Answers:       []int{1},
		QuestionMedia: "text",
		OptionsMedia:  "text",
		Resource:      EmptySQLNullString(),
	}

	payload := BuildLiveScoreboardPayload(question, nil, nil, 5, 20, false)
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := security.AssertHasSensitiveAnswerKeyField(body, "answers"); err != nil {
		t.Fatalf("scoreboard payload should include answers after question closes: %v", err)
	}
}
