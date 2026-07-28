package security

import (
	"testing"
)

func TestFindSensitiveAnswerKeyFields(t *testing.T) {
	payload := map[string]any{
		"data": []any{
			map[string]any{
				"question": "Stem",
				"nested": map[string]any{
					"official_answer": "[1]",
				},
			},
		},
	}

	fields := FindSensitiveAnswerKeyFields(payload)
	if len(fields) != 1 || fields[0] != "official_answer" {
		t.Fatalf("fields = %#v, want [official_answer]", fields)
	}
}

func TestAssertNoSensitiveAnswerKeyFields_passesCleanPayload(t *testing.T) {
	body := []byte(`{"data":{"question":"Stem","options":{"1":"A"}}}`)
	if err := AssertNoSensitiveAnswerKeyFields(body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssertNoSensitiveAnswerKeyFields_failsOnLeak(t *testing.T) {
	body := []byte(`{"data":{"correct_answer":"[1]"}}`)
	if err := AssertNoSensitiveAnswerKeyFields(body); err == nil {
		t.Fatal("expected leak detection error")
	}
}

func TestAssertHasSensitiveAnswerKeyField(t *testing.T) {
	body := []byte(`{"data":[{"correct_answer":"[2]"}]}`)
	if err := AssertHasSensitiveAnswerKeyField(body, "correct_answer"); err != nil {
		t.Fatalf("expected correct_answer present: %v", err)
	}
}
