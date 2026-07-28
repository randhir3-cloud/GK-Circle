package models

import (
	"testing"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func TestApplyAnswerAuthorityDefaults(t *testing.T) {
	authority, operational, err := ApplyAnswerAuthority(
		[]int{2},
		nil,
		nil,
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(operational) != 1 || operational[0] != 2 {
		t.Fatalf("operational answers: got %v", operational)
	}
	if len(authority.OfficialAnswer) != 1 || authority.OfficialAnswer[0] != 2 {
		t.Fatalf("official answer: got %v", authority.OfficialAnswer)
	}
	if authority.AnswerReviewStatus != constants.AnswerReviewUnreviewed {
		t.Fatalf("status: got %s", authority.AnswerReviewStatus)
	}
}

func TestApplyAnswerAuthorityExplicitOfficial(t *testing.T) {
	authority, operational, err := ApplyAnswerAuthority(
		[]int{1},
		[]int{3},
		[]int{2},
		constants.AnswerReviewRevised,
		"key corrected",
		"exam notice",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if operational[0] != 2 {
		t.Fatalf("operational: got %v", operational)
	}
	if authority.OfficialAnswer[0] != 3 {
		t.Fatalf("official: got %v", authority.OfficialAnswer)
	}
	if authority.AnswerReviewStatus != constants.AnswerReviewRevised {
		t.Fatalf("status: got %s", authority.AnswerReviewStatus)
	}
	if authority.AnswerRevisionReason != "key corrected" {
		t.Fatalf("reason: got %s", authority.AnswerRevisionReason)
	}
}

func TestApplyAnswerAuthorityInvalidStatus(t *testing.T) {
	_, _, err := ApplyAnswerAuthority([]int{1}, nil, nil, "INVALID", "", "")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}
