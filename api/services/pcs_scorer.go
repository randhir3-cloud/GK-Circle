package services

import (
	"math"
	"sort"

	"github.com/google/uuid"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
	"github.com/randhir3-cloud/GK-Circle-v2/api/models"
)

// ScoringKeySource identifies which frozen snapshot field supplied the key.
const (
	ScoringKeyAuthoritative = "AUTHORITATIVE"
	ScoringKeyOperational   = "OPERATIONAL"
	ScoringKeyNone          = "NONE"
)

// PCSScoreConfig is quiz-level mark configuration applied at submit.
type PCSScoreConfig struct {
	NegativeMarksPerQuestion float64
}

type PCSQuestionScoreInput struct {
	QuestionID      uuid.UUID
	Type            int
	Points          *int16
	SelectedOptions []int
	// Frozen keys from the snapshot item only.
	AuthoritativeAnswer []int
	OperationalAnswers  []int
}

type PCSQuestionScoreResult struct {
	QuestionID      uuid.UUID
	SelectedOptions []int
	IsCorrect       *bool
	Score           float64
	MaxPoints       float64
	Scored          bool // false for survey/unscored
	KeySource       string
}

type PCSAttemptScoreResult struct {
	QuestionResults []PCSQuestionScoreResult
	TotalScore      float64
	MaxScore        float64
	CorrectCount    int
	IncorrectCount  int
	UnansweredCount int
	UnscoredCount   int
}

// ScorePCSAttempt calculates deterministic mark-based outcomes from frozen snapshot data.
func ScorePCSAttempt(items []PCSQuestionScoreInput, cfg PCSScoreConfig) PCSAttemptScoreResult {
	out := PCSAttemptScoreResult{
		QuestionResults: make([]PCSQuestionScoreResult, 0, len(items)),
	}
	for _, item := range items {
		result := scorePCSQuestion(item, cfg)
		out.QuestionResults = append(out.QuestionResults, result)
		out.MaxScore += result.MaxPoints
		out.TotalScore += result.Score
		if !result.Scored {
			out.UnscoredCount++
			continue
		}
		if len(result.SelectedOptions) == 0 {
			out.UnansweredCount++
			continue
		}
		if result.IsCorrect != nil && *result.IsCorrect {
			out.CorrectCount++
		} else {
			out.IncorrectCount++
		}
	}
	out.TotalScore = floorScoreAtZero(roundScore2(out.TotalScore))
	out.MaxScore = roundScore2(out.MaxScore)
	return out
}

func scorePCSQuestion(item PCSQuestionScoreInput, cfg PCSScoreConfig) PCSQuestionScoreResult {
	points := questionPoints(item.Points)
	result := PCSQuestionScoreResult{
		QuestionID:      item.QuestionID,
		SelectedOptions: uniqueSortedCopy(item.SelectedOptions),
		MaxPoints:       points,
		Score:           0,
		Scored:          true,
	}

	if item.Type == constants.Survey {
		result.Scored = false
		result.MaxPoints = 0
		result.IsCorrect = nil
		result.KeySource = ScoringKeyNone
		return result
	}

	key, source := resolveFrozenScoringKey(item.AuthoritativeAnswer, item.OperationalAnswers)
	result.KeySource = source
	if source == ScoringKeyNone {
		// No frozen key: treat as unscored to avoid inventing correctness.
		result.Scored = false
		result.MaxPoints = 0
		result.IsCorrect = nil
		return result
	}

	if len(result.SelectedOptions) == 0 {
		result.IsCorrect = nil
		result.Score = 0
		return result
	}

	correct := equalIntSets(result.SelectedOptions, key)
	result.IsCorrect = &correct
	if correct {
		result.Score = roundScore2(points)
	} else {
		result.Score = roundScore2(-cfg.NegativeMarksPerQuestion)
	}
	return result
}

func resolveFrozenScoringKey(authoritative, operational []int) ([]int, string) {
	if len(authoritative) > 0 {
		return uniqueSortedCopy(authoritative), ScoringKeyAuthoritative
	}
	if len(operational) > 0 {
		return uniqueSortedCopy(operational), ScoringKeyOperational
	}
	return nil, ScoringKeyNone
}

func questionPoints(points *int16) float64 {
	if points == nil {
		return 1
	}
	if *points < 0 {
		return 0
	}
	return float64(*points)
}

func equalIntSets(a, b []int) bool {
	left := uniqueSortedCopy(a)
	right := uniqueSortedCopy(b)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func uniqueSortedCopy(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(values))
	out := make([]int, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func roundScore2(value float64) float64 {
	return math.Round(value*100) / 100
}

func floorScoreAtZero(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

// SnapshotItemToScoreInput maps a frozen snapshot item + selected options into scorer input.
func SnapshotItemToScoreInput(item models.TestSnapshotItem, selected []int) PCSQuestionScoreInput {
	return PCSQuestionScoreInput{
		QuestionID:          item.QuestionID,
		Type:                item.Type,
		Points:              item.Points,
		SelectedOptions:     selected,
		AuthoritativeAnswer: item.AuthoritativeAnswer,
		OperationalAnswers:  item.Answers,
	}
}

// AttemptSnapshotItemToScoreInput maps attempt-linked freeze rows into scorer input.
func AttemptSnapshotItemToScoreInput(item models.AssessmentAttemptSnapshotItem, selected []int) PCSQuestionScoreInput {
	return PCSQuestionScoreInput{
		QuestionID:          item.QuestionID,
		Type:                item.Type,
		Points:              item.Points,
		SelectedOptions:     selected,
		AuthoritativeAnswer: item.AuthoritativeAnswer,
		OperationalAnswers:  item.Answers,
	}
}

// ExpectedMaxScoreFromSnapshotItems sums points for scored (non-survey) freeze items.
func ExpectedMaxScoreFromSnapshotItems(items []models.CreateAttemptSnapshotItemParams) float64 {
	var total float64
	for _, item := range items {
		if item.Type == constants.Survey {
			continue
		}
		total += questionPoints(item.Points)
	}
	return roundScore2(total)
}
