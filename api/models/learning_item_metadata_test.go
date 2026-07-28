package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestNormalizeLearningItemMetadataDefaultsAndValid(t *testing.T) {
	t.Run("omit defaults to version 1 empty blocks", func(t *testing.T) {
		got, err := normalizeLearningItemMetadata(nil)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		if !bytes.Equal(got, defaultLearningItemMetadata) {
			t.Fatalf("got %s, want %s", got, defaultLearningItemMetadata)
		}
	})

	t.Run("empty blocks allowed", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[]}`)
		got, err := normalizeLearningItemMetadata(input)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		var decoded LearningItemMetadata
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Version != 1 || decoded.Blocks == nil || len(decoded.Blocks) != 0 {
			t.Fatalf("unexpected metadata: %+v", decoded)
		}
	})

	t.Run("preserves non-empty data object", func(t *testing.T) {
		input := json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{"text":"hello","extra":1}}]}`)
		got, err := normalizeLearningItemMetadata(input)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		var decoded LearningItemMetadata
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(decoded.Blocks) != 1 {
			t.Fatalf("blocks = %d", len(decoded.Blocks))
		}
		if !bytes.Equal(bytes.TrimSpace(decoded.Blocks[0].Data), []byte(`{"text":"hello","extra":1}`)) {
			t.Fatalf("data coerced or rewritten: %s", decoded.Blocks[0].Data)
		}
		if bytes.Equal(bytes.TrimSpace(decoded.Blocks[0].Data), []byte(`{}`)) {
			t.Fatal("data was coerced to empty object")
		}
	})
}

func TestNormalizeLearningItemMetadataRejects(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
		want  error
	}{
		{name: "null root", input: json.RawMessage(`null`), want: ErrLearningItemMetadataInvalid},
		{name: "array root", input: json.RawMessage(`[]`), want: ErrLearningItemMetadataInvalid},
		{name: "missing blocks", input: json.RawMessage(`{"version":1}`), want: ErrLearningItemMetadataInvalid},
		{name: "null blocks", input: json.RawMessage(`{"version":1,"blocks":null}`), want: ErrLearningItemMetadataInvalid},
		{name: "version zero", input: json.RawMessage(`{"version":0,"blocks":[]}`), want: ErrLearningItemMetadataVersionInvalid},
		{name: "version missing", input: json.RawMessage(`{"blocks":[]}`), want: ErrLearningItemMetadataVersionInvalid},
		{name: "duplicate ids", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":{}},{"id":"b1","type":"HEADING","data":{}}]}`), want: ErrLearningItemBlockDuplicate},
		{name: "blank id", input: json.RawMessage(`{"version":1,"blocks":[{"id":"  ","type":"TEXT","data":{}}]}`), want: ErrLearningItemMetadataInvalid},
		{name: "unsupported type", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"WIDGET","data":{}}]}`), want: ErrLearningItemBlockTypeInvalid},
		{name: "data null", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":null}]}`), want: ErrLearningItemMetadataInvalid},
		{name: "data missing", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT"}]}`), want: ErrLearningItemMetadataInvalid},
		{name: "data array", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":[]}]}`), want: ErrLearningItemMetadataInvalid},
		{name: "data scalar", input: json.RawMessage(`{"version":1,"blocks":[{"id":"b1","type":"TEXT","data":"x"}]}`), want: ErrLearningItemMetadataInvalid},
		{name: "empty object root", input: json.RawMessage(`{}`), want: ErrLearningItemMetadataInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := normalizeLearningItemMetadata(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if err != nil && (containsSQLLeak(err.Error())) {
				t.Fatalf("error leaked parse details: %v", err)
			}
		})
	}
}

func containsSQLLeak(message string) bool {
	for _, needle := range []string{"pq:", "SQLSTATE", "json:", "unmarshal"} {
		if bytes.Contains([]byte(message), []byte(needle)) {
			return true
		}
	}
	return false
}
