package security

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SensitiveAnswerKeyFields are JSON object keys that must not appear in
// pre-authorised or pre-review HTTP/WebSocket payloads.
var SensitiveAnswerKeyFields = []string{
	"answers",
	"correct_answer",
	"official_answer",
	"authoritative_answer",
	"is_correct",
}

var sensitiveAnswerKeySet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(SensitiveAnswerKeyFields))
	for _, field := range SensitiveAnswerKeyFields {
		set[field] = struct{}{}
	}
	return set
}()

// FindSensitiveAnswerKeyFields walks decoded JSON and returns any sensitive
// answer-key field names present at any nesting depth.
func FindSensitiveAnswerKeyFields(value any) []string {
	found := make(map[string]struct{})
	collectSensitiveAnswerKeyFields(value, found)

	fields := make([]string, 0, len(found))
	for field := range found {
		fields = append(fields, field)
	}
	return fields
}

func collectSensitiveAnswerKeyFields(value any, found map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if _, ok := sensitiveAnswerKeySet[key]; ok {
				found[key] = struct{}{}
			}
			collectSensitiveAnswerKeyFields(nested, found)
		}
	case []any:
		for _, item := range typed {
			collectSensitiveAnswerKeyFields(item, found)
		}
	}
}

// AssertNoSensitiveAnswerKeyFields fails when any sensitive answer-key field is
// present in the JSON document.
func AssertNoSensitiveAnswerKeyFields(body []byte) error {
	if len(body) == 0 {
		return nil
	}

	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	fields := FindSensitiveAnswerKeyFields(decoded)
	if len(fields) == 0 {
		return nil
	}
	return fmt.Errorf("unexpected sensitive answer-key fields: %s", strings.Join(fields, ", "))
}

// AssertHasSensitiveAnswerKeyField fails when the named field is absent from JSON.
func AssertHasSensitiveAnswerKeyField(body []byte, field string) error {
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	fields := FindSensitiveAnswerKeyFields(decoded)
	for _, present := range fields {
		if present == field {
			return nil
		}
	}
	return fmt.Errorf("expected field %q in response payload", field)
}
