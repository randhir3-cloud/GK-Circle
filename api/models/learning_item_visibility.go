package models

import (
	"bytes"
	"encoding/json"
)

type LearningItemVisibilityMode string

const (
	LearningItemVisibilityModeAll           LearningItemVisibilityMode = "ALL"
	LearningItemVisibilityModeAuthenticated LearningItemVisibilityMode = "AUTHENTICATED"
	LearningItemVisibilityModePremium       LearningItemVisibilityMode = "PREMIUM"
	LearningItemVisibilityModeInstructor    LearningItemVisibilityMode = "INSTRUCTOR"
	LearningItemVisibilityModeHidden        LearningItemVisibilityMode = "HIDDEN"
)

// LearningItemBlockVisibility is persisted metadata. Write-time validation lives
// here (COURSE-P2-T05); runtime learner projection is in
// learning_item_visibility_runtime.go (COURSE-P2-T13).
type LearningItemBlockVisibility struct {
	Mode LearningItemVisibilityMode `json:"mode"`
}

var defaultLearningItemBlockVisibility = LearningItemBlockVisibility{
	Mode: LearningItemVisibilityModeAll,
}

// normalizeLearningItemBlockVisibility distinguishes omitted vs null visibility.
// Omitted (empty raw) defaults to ALL. Null/non-object/missing mode are rejected.
func normalizeLearningItemBlockVisibility(raw json.RawMessage) (LearningItemBlockVisibility, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return defaultLearningItemBlockVisibility, nil
	}
	trimmed := bytes.TrimSpace(raw)
	if string(trimmed) == "null" {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityInvalid
	}
	if !json.Valid(trimmed) {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityInvalid
	}

	var probe interface{}
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityInvalid
	}
	object, ok := probe.(map[string]interface{})
	if !ok {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityInvalid
	}
	modeValue, hasMode := object["mode"]
	if !hasMode {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityInvalid
	}
	modeString, ok := modeValue.(string)
	if !ok {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityModeInvalid
	}
	mode := LearningItemVisibilityMode(modeString)
	if !isValidLearningItemVisibilityMode(mode) {
		return LearningItemBlockVisibility{}, ErrLearningItemVisibilityModeInvalid
	}
	return LearningItemBlockVisibility{Mode: mode}, nil
}

func isValidLearningItemVisibilityMode(mode LearningItemVisibilityMode) bool {
	switch mode {
	case LearningItemVisibilityModeAll,
		LearningItemVisibilityModeAuthenticated,
		LearningItemVisibilityModePremium,
		LearningItemVisibilityModeInstructor,
		LearningItemVisibilityModeHidden:
		return true
	default:
		return false
	}
}
