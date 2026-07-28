package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestValidateLearningItemPlaceholderString(t *testing.T) {
	valid := []string{
		"plain text",
		"Welcome {{student_name}}",
		"Hi {{student_name}} and {{student_name}} again",
		"{{course_title}} — {{lesson1}}",
		"brace { ordinary } text",
		"{{" + strings.Repeat("a", 64) + "}}",
	}
	for _, text := range valid {
		if err := validateLearningItemPlaceholderString(text); err != nil {
			t.Fatalf("valid %q: %v", text, err)
		}
	}

	tests := []struct {
		name string
		text string
		want error
	}{
		{name: "empty", text: "{{}}", want: ErrLearningItemPlaceholderSyntax},
		{name: "leading digit", text: "{{123name}}", want: ErrLearningItemPlaceholderInvalid},
		{name: "hyphen", text: "{{student-name}}", want: ErrLearningItemPlaceholderInvalid},
		{name: "spaces", text: "{{ student }}", want: ErrLearningItemPlaceholderInvalid},
		{name: "punctuation", text: "{{name!}}", want: ErrLearningItemPlaceholderInvalid},
		{name: "unclosed", text: "Hello {{name", want: ErrLearningItemPlaceholderSyntax},
		{name: "nested open", text: "{{{name}}}", want: ErrLearningItemPlaceholderSyntax},
		{name: "too long", text: "{{" + strings.Repeat("a", 65) + "}}", want: ErrLearningItemPlaceholderInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateLearningItemPlaceholderString(test.text)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNormalizeLearningItemMetadataPlaceholders(t *testing.T) {
	t.Run("nested and array strings valid", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Hi {{student_name}}","meta":{"title":"{{course_title}}"},"tags":["{{lesson1}}","plain"]}}]}`)
		got, err := normalizeLearningItemMetadata(input)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		var decoded LearningItemMetadata
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !bytes.Contains(decoded.Blocks[0].Data, []byte(`{{student_name}}`)) {
			t.Fatalf("placeholder rewritten: %s", decoded.Blocks[0].Data)
		}
	})

	t.Run("invalid placeholder rejected", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Hi {{student-name}}"}}]}`)
		_, err := normalizeLearningItemMetadata(input)
		if !errors.Is(err, ErrLearningItemPlaceholderInvalid) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemPlaceholderInvalid)
		}
	})

	t.Run("syntax error rejected", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"Hi {{name"}}]}`)
		_, err := normalizeLearningItemMetadata(input)
		if !errors.Is(err, ErrLearningItemPlaceholderSyntax) {
			t.Fatalf("error = %v, want %v", err, ErrLearningItemPlaceholderSyntax)
		}
	})
}
